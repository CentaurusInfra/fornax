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
	"fmt"
	"reflect"
	"time"

	edgeclustersv1 "github.com/kubeedge/kubeedge/cloud/pkg/apis/edgeclusters/v1"
	crdClientset "github.com/kubeedge/kubeedge/cloud/pkg/client/clientset/versioned"
	crdinformers "github.com/kubeedge/kubeedge/cloud/pkg/client/informers/externalversions"
	ecinformers "github.com/kubeedge/kubeedge/cloud/pkg/client/informers/externalversions/edgeclusters/v1"
	eclisters "github.com/kubeedge/kubeedge/cloud/pkg/client/listers/edgeclusters/v1"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/edgecore/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	cache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/apimachinery/pkg/api/errors"
	labels "k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type MissionStatusReporter struct {
	ClusterName                 string
	KubeconfigFile              string
	CrdClient                   *crdClientset.Clientset
	MissionInformer             ecinformers.MissionInformer
	MissionLister               eclisters.MissionLister
	EdgeclusterLister           eclisters.EdgeClusterLister
	MissionInformerSynced       cache.InformerSynced
	EdgeclusterListerSynced     cache.InformerSynced
	resyncPeriod                time.Duration
	queue                       workqueue.RateLimitingInterface
	statusCache                 map[string]map[string]string
	statusSyncer                func(key string) error
	clusterd                    *clusterd
	missionDeployer             *MissionDeployer
	missionStatusUpdateInterval time.Duration
}

//NewMissionStatusReporter creates new mission status  object
func NewMissionStatusReporter(clusterdConfig *v1alpha1.Clusterd, c *clusterd, md *MissionDeployer, stopCh <-chan struct{}) *MissionStatusReporter {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", clusterdConfig.Kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to build config, err: %v", err)
	}

	crdKubeConfig := rest.CopyConfig(kubeConfig)
	crdKubeConfig.ContentType = runtime.ContentTypeJSON

	crdClient := crdClientset.NewForConfigOrDie(crdKubeConfig)

	resyncPeriod := time.Duration(clusterdConfig.InformerResyncInterval) * time.Second
	crdInformerFactory := crdinformers.NewSharedInformerFactory(crdClient, resyncPeriod)
	missionInformer := crdInformerFactory.Edgeclusters().V1().Missions()
	edgeclusterInformer := crdInformerFactory.Edgeclusters().V1().EdgeClusters()

	msw := &MissionStatusReporter{
		ClusterName:                 clusterdConfig.Name,
		KubeconfigFile:              clusterdConfig.Kubeconfig,
		CrdClient:                   crdClient,
		MissionInformer:             missionInformer,
		MissionLister:               missionInformer.Lister(),
		EdgeclusterLister:           edgeclusterInformer.Lister(),
		statusCache:                 map[string]map[string]string{},
		queue:                       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "mission"),
		clusterd:                    c,
		missionDeployer:             md,
		missionStatusUpdateInterval: time.Duration(clusterdConfig.MissionStatusUpdateInterval) * time.Second,
		resyncPeriod:                resyncPeriod,
	}

	missionInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				msw.enqueue(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				newMission, ok1 := newObj.(*edgeclustersv1.Mission)
				oldMission, ok2 := oldObj.(*edgeclustersv1.Mission)
				if !ok1 || !ok2 {
					klog.Errorf("obj is not of Missiion type")
					return
				}
				// we only update if there is difference in status
				// difference in spec will be handled by clusterd itself
				if !reflect.DeepEqual(oldMission.Status, newMission.Status) {
					klog.V(4).Infof("\n there is  diff old status (%v) new status (%v)", oldMission.Status, newMission.Status)
					msw.enqueue(newObj)
				}
			},
		},
		resyncPeriod,
	)

	msw.MissionInformerSynced = missionInformer.Informer().HasSynced
	msw.EdgeclusterListerSynced = edgeclusterInformer.Informer().HasSynced

	go missionInformer.Informer().Run(stopCh)
	go edgeclusterInformer.Informer().Run(stopCh)

	msw.statusSyncer = msw.syncMissionStatus

	return msw
}

// enqueue adds an object to the work queue
func (m *MissionStatusReporter) enqueue(obj interface{}) {
	mission, ok := obj.(*edgeclustersv1.Mission)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("Not a mission object: %v", obj))
		return
	}

	m.queue.Add(mission.Name)
}

// Run starts the  with the specified number of workers.
func (m *MissionStatusReporter) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer m.queue.ShutDown()

	klog.Infof("Starting mission status .")
	defer klog.Infof("Shutting down  mission status ")

	klog.V(3).Infof("Starting syncing mission & edgecluster informer")
	cache.WaitForCacheSync(stopCh, m.MissionInformerSynced)
	cache.WaitForCacheSync(stopCh, m.EdgeclusterListerSynced)
	klog.V(3).Infof("informer synced")

	klog.V(5).Info("Starting workers of mission controller")
	for i := 0; i < workers; i++ {
		go utilwait.Until(m.worker, m.resyncPeriod, stopCh)
	}

	go utilwait.Until(m.localMissionStatusUpdate, m.missionStatusUpdateInterval, stopCh)
	<-stopCh
}

// worker processes the queue of mission objects.
func (m *MissionStatusReporter) worker() {
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
func (m *MissionStatusReporter) processQueue(missionName string) (err error) {
	klog.V(3).Infof("Starting processsing queue for mission: %v", missionName)

	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing mission %q (%v)", missionName, time.Since(startTime))
	}()

	_, err = m.MissionLister.Get(missionName)
	if errors.IsNotFound(err) {
		klog.Infof("mission has been deleted %v", missionName)
		return nil
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Unable to retrieve mission %v from store: %v", missionName, err))
		return err
	}

	return m.statusSyncer(missionName)
}

func (m *MissionStatusReporter) syncMissionStatus(missionName string) (err error) {
	klog.V(3).Infof("Starting syncing status of mission: %v", missionName)

	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing status of mission %q (%v)", missionName, time.Since(startTime))
	}()

	updated, err := m.updateStatusCache(missionName)
	if err != nil {
		return err
	}

	if !updated {
		return nil
	}

	return m.clusterd.UpdateMissionStatus(missionName, m.statusCache[missionName])
}

func (m *MissionStatusReporter) updateStatusCache(missionName string) (updated bool, err error) {
	// no error as its caller, processQueue, has checked.
	mission, _ := m.MissionLister.Get(missionName)

	val, found := m.statusCache[missionName]

	if !found || !reflect.DeepEqual(val, mission.Status) {
		klog.V(4).Infof("Detected difference: (%v) (%v)", val, mission.Status)
		m.statusCache[missionName] = mission.Status
		return true, nil
	}

	return false, nil
}

func (m *MissionStatusReporter) localMissionStatusUpdate() {
	klog.V(4).Infof("Start checking the status of missions...")
	missions, err := m.MissionLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("Error listing missions: %v", err)
	}
	for _, mission := range missions {
		m.missionDeployer.StatusUpdate(mission, true)
	}
}
