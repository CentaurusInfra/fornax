/*
Copyright 2015 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package clusterd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	edgeclustersv1 "github.com/kubeedge/kubeedge/cloud/pkg/apis/edgeclusters/v1"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/edgecore/v1alpha1"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	// it forces the state of each mission is reported to upper layer at least once per minute.
	// We set this value hard-coded, instead of user-configurable, as it will be a performance if it is not set too small.
	ForcedResyncInterval = 60
)

type MissionStateReporter struct {
	ClusterName                string
	KubeconfigFile             string
	resyncPeriod               time.Duration
	queue                      workqueue.RateLimitingInterface
	missionCache               map[string]edgeclustersv1.Mission
	clusterd                   *clusterd
	missionDeployer            *MissionDeployer
	missionStateUpdateInterval time.Duration
	KubectlCli                 string
	MaxStateIdleCycles         int
	StateIdleCycles            map[string]int
}

//NewMissionStateReporter creates new mission state  object
func NewMissionStateReporter(clusterdConfig *v1alpha1.Clusterd, c *clusterd, md *MissionDeployer, stopCh <-chan struct{}) *MissionStateReporter {

	resyncPeriod := time.Duration(clusterdConfig.ResyncInterval) * time.Second
	basedir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	msw := &MissionStateReporter{
		ClusterName:                clusterdConfig.Name,
		KubeconfigFile:             clusterdConfig.Kubeconfig,
		missionCache:               map[string]edgeclustersv1.Mission{},
		queue:                      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "mission"),
		clusterd:                   c,
		missionDeployer:            md,
		missionStateUpdateInterval: time.Duration(clusterdConfig.MissionStateUpdateInterval) * time.Second,
		resyncPeriod:               resyncPeriod,
		KubectlCli:                 filepath.Join(basedir, DistroToKubectl[clusterdConfig.KubeDistro]),
		MaxStateIdleCycles:         ForcedResyncInterval / int(clusterdConfig.ResyncInterval),
		StateIdleCycles:            map[string]int{},
	}

	return msw
}

// Run starts the  with the specified number of workers.
func (m *MissionStateReporter) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer m.queue.ShutDown()

	klog.Infof("Starting mission state reporter.")
	defer klog.Infof("Shutting down  mission state reporter")

	klog.V(5).Info("Starting the state syncer in the mission state reporter")
	go utilwait.Until(m.stateSyncer, m.resyncPeriod, stopCh)

	klog.V(5).Info("Starting workers of mission state reporter")
	for i := 0; i < workers; i++ {
		go utilwait.Until(m.worker, m.resyncPeriod, stopCh)
	}

	go utilwait.Until(m.localMissionStateUpdate, m.missionStateUpdateInterval, stopCh)
	<-stopCh
}

func (m *MissionStateReporter) stateSyncer() {
	get_mission_cmd := fmt.Sprintf(" %s get missions -o json --kubeconfig=%s | jq .items ", m.KubectlCli, m.KubeconfigFile)
	output, err := ExecCommandLine(get_mission_cmd, COMMAND_TIMEOUT_SEC)
	if err != nil {
		klog.Errorf("Failed to get mission: %v", err)
		return
	}

	var missionList []edgeclustersv1.Mission
	if err = json.Unmarshal([]byte(output), &missionList); err != nil {
		klog.Errorf("Failed to unmarshal mission list: %v", err)
	}

	newmissionCache := map[string]edgeclustersv1.Mission{}
	for _, mission := range missionList {
		newmissionCache[mission.Name] = mission
		_, exists := m.missionCache[mission.Name]
		// if the mission state changes, we send an update
		if !exists || !EqualMaps(m.missionCache[mission.Name].State, mission.State) {
			m.queue.Add(mission.Name)
		} else {
			// if the state has been idle for a long time, we send an update
			if i, ok := m.StateIdleCycles[mission.Name]; ok {
				m.StateIdleCycles[mission.Name] += 1
				if i > m.MaxStateIdleCycles {
					m.queue.Add(mission.Name)
				}
			}
		}
	}

	m.missionCache = newmissionCache
}

// worker processes the queue of mission objects.
func (m *MissionStateReporter) worker() {
	workFunc := func() bool {
		key, quit := m.queue.Get()
		if quit {
			return true
		}
		defer m.queue.Done(key)

		err := m.processQueue(key.(string))
		if err == nil {
			// no error, forget this entry and return
			m.queue.Forget(key)
			return false
		} else {
			// rather than wait for a full resync, re-add the mission to the queue to be processed
			m.queue.AddRateLimited(key)
			utilruntime.HandleError(err)
		}
		return false
	}

	for {
		quit := workFunc()
		if quit {
			return
		}
	}
}

// processQueue looks for a mission with the specified name and synchronizes it
func (m *MissionStateReporter) processQueue(missionName string) (err error) {
	klog.V(3).Infof("Starting processsing queue for mission: %v", missionName)

	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing mission %q (%v)", missionName, time.Since(startTime))
	}()

	// reset the value of state idle cycles for this mission
	m.StateIdleCycles[missionName] = 0

	return m.clusterd.UpdateMissionState(missionName, m.missionCache[missionName].State)
}

func (m *MissionStateReporter) syncMissionState(missionName string) (err error) {
	klog.V(3).Infof("Starting syncing state of mission: %v", missionName)

	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing state of mission %q (%v)", missionName, time.Since(startTime))
	}()

	return m.clusterd.UpdateMissionState(missionName, m.missionCache[missionName].State)
}

func (m *MissionStateReporter) localMissionStateUpdate() {
	klog.V(4).Infof("Start checking the state of missions...")
	for _, mission := range m.missionCache {
		m.missionDeployer.StateUpdate(&mission, true)
	}
}
