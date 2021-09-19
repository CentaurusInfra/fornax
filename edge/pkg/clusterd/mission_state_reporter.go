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
	"strings"
	"sync"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	edgeclustersv1 "github.com/kubeedge/kubeedge/cloud/pkg/apis/edgeclusters/v1"
	edgeapi "github.com/kubeedge/kubeedge/common/types"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/config"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/helper"
)

const (
	// it forces the state of each mission is reported to upper layer at least once per minute.
	// We set this value hard-coded, instead of user-configurable, as it will be a performance if it is not set too small.
	ForcedResyncInterval = 60
	LocalEdgeCluster     = "LocalEdgeCluster"
	StatusNotMatch       = "not match"
)

var stateCycleLock sync.Mutex

type MissionStateReporter struct {
	resyncPeriod               time.Duration
	queue                      workqueue.RateLimitingInterface
	missionCache               map[string]edgeclustersv1.Mission
	clusterd                   *clusterd
	missionDeployer            *MissionDeployer
	missionStateUpdateInterval time.Duration
	maxStateIdleCycles         int
	stateIdleCycles            map[string]int
}

//NewMissionStateReporter creates new mission state  object
func NewMissionStateReporter(c *clusterd, md *MissionDeployer) *MissionStateReporter {
	return &MissionStateReporter{
		missionCache:               map[string]edgeclustersv1.Mission{},
		queue:                      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "mission"),
		clusterd:                   c,
		missionDeployer:            md,
		missionStateUpdateInterval: time.Duration(config.Config.MissionStateUpdateInterval) * time.Second,
		resyncPeriod:               time.Duration(config.Config.ResyncInterval) * time.Second,
		maxStateIdleCycles:         ForcedResyncInterval / int(config.Config.ResyncInterval),
		stateIdleCycles:            map[string]int{},
	}
}

// Run starts the  with the specified number of workers.
func (m *MissionStateReporter) Run() {
	defer utilruntime.HandleCrash()
	defer m.queue.ShutDown()

	stopChan := make(chan struct{})

	klog.Infof("Starting mission state reporter.")
	defer klog.Infof("Shutting down  mission state reporter")

	klog.V(4).Info("Starting the state syncer in the mission state reporter")
	go utilwait.Until(m.stateSyncer, m.resyncPeriod, utilwait.NeverStop)

	klog.V(4).Info("Starting workers of mission state reporter")
	for i := 0; i < config.Config.MissionStateWatchWorkers; i++ {
		go utilwait.Until(m.syncQueue, m.resyncPeriod, stopChan)
	}

	go utilwait.Until(m.updateLocalMissionState, m.missionStateUpdateInterval, utilwait.NeverStop)
	<-stopChan
}

func (m *MissionStateReporter) stateSyncer() {
	getMissionCmd := fmt.Sprintf(" %s get missions -o json --kubeconfig=%s | jq .items ", config.Config.KubectlCli, config.Config.Kubeconfig)
	output, err := helper.ExecCommandToCluster(getMissionCmd)
	if err != nil {
		klog.Errorf("Failed to get mission: %v", err)
		return
	}

	if strings.Contains(output, "the server could not find the requested resource") {
		klog.Infof("Looks like the mission CRD is not applied in the edge cluster yet, we got the error: %v", output)
		return
	}

	var missionList []edgeclustersv1.Mission

	if err = json.Unmarshal([]byte(output), &missionList); err != nil {
		klog.Errorf("Failed to unmarshal mission list: %v, output : (%v)", err, output)
	}

	newmissionCache := map[string]edgeclustersv1.Mission{}
	for _, mission := range missionList {
		newmissionCache[mission.Name] = mission
		_, exists := m.missionCache[mission.Name]
		// if the mission state changes, we send an update
		if !exists || !helper.EqualMaps(m.missionCache[mission.Name].State, mission.State) {
			m.queue.Add(mission.Name)
		} else {
			// if the state has been idle for a long time, we send an update
			stateCycleLock.Lock()
			if missionIdleCycles, exists := m.stateIdleCycles[mission.Name]; exists {
				m.stateIdleCycles[mission.Name]++

				if missionIdleCycles > m.maxStateIdleCycles {
					m.queue.Add(mission.Name)
				}
			}
			stateCycleLock.Unlock()
		}
	}

	m.missionCache = newmissionCache
}

// syncQueue processes the queue of mission objects.
func (m *MissionStateReporter) syncQueue() {
	workFunc := func() bool {
		if !helper.TestClusterReady() {
			klog.V(3).Infof("Cluster is unhealthy, skipping mission status checking.")
			return false
		}

		key, quit := m.queue.Get()
		if quit {
			return true
		}
		defer m.queue.Done(key)

		err := m.processQueueItem(key.(string))
		if err == nil {
			// no error, forget this entry and return
			m.queue.Forget(key)
			return false
		}

		// rather than wait for a full resync, re-add the mission to the queue to be processed
		m.queue.AddRateLimited(key)
		utilruntime.HandleError(err)
		return false
	}

	for {
		quit := workFunc()
		if quit {
			return
		}
	}
}

// processQueueItem looks for a mission with the specified name and synchronizes it
func (m *MissionStateReporter) processQueueItem(missionName string) (err error) {
	klog.V(3).Infof("Starting processsing queue for mission: %v", missionName)

	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing mission %q (%v)", missionName, time.Since(startTime))
	}()

	// reset the value of state idle cycles for this mission
	stateCycleLock.Lock()
	m.stateIdleCycles[missionName] = 0
	stateCycleLock.Unlock()

	return m.ReportMissionState(missionName, m.missionCache[missionName].State)
}

func (m *MissionStateReporter) syncMissionState(missionName string) (err error) {
	klog.V(3).Infof("Starting syncing state of mission: %v", missionName)

	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing state of mission %q (%v)", missionName, time.Since(startTime))
	}()

	return m.ReportMissionState(missionName, m.missionCache[missionName].State)
}

func (m *MissionStateReporter) updateLocalMissionState() {
	klog.V(4).Infof("Start checking the state of missions...")
	for _, mission := range m.missionCache {
		m.missionDeployer.UpdateState(&mission, true)
	}
}

func (m *MissionStateReporter) ReportMissionState(missionName string, missionState map[string]string) error {
	updatedMissionState := map[string]string{}
	clusterName := config.Config.Name
	for key, val := range missionState {
		if key == LocalEdgeCluster {
			updatedMissionState[clusterName] = val
		} else {
			updatedMissionState[clusterName+"/"+key] = val
		}
	}
	msRequest := edgeapi.MissionStateRequest{
		UID:         m.clusterd.uid,
		ClusterName: clusterName,
		State:       updatedMissionState,
	}

	err := m.clusterd.metaClient.MissionState(m.clusterd.namespace).Update(missionName, msRequest)
	if err != nil {
		klog.Errorf("update mission %v status failed, error: %v", missionName, err)
	}

	return err
}
