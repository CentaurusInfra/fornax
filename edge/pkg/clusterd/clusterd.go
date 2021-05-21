/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

@CHANGELOG
KubeEdge Authors: To create mini-kubelet for edge deployment scenario,
This file is derived from K8S Kubelet code with reduced set of methods
Changes done are
1. package clusterd got some functions from "k8s.io/kubernetes/pkg/kubelet/kubelet.go"
and made some variant
*/

package clusterd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kubeedge/beehive/pkg/common/util"
	"github.com/kubeedge/beehive/pkg/core"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/beehive/pkg/core/model"
	edgeclustersv1 "github.com/kubeedge/kubeedge/cloud/pkg/apis/edgeclusters/v1"
	"github.com/kubeedge/kubeedge/common/constants"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/config"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
	"github.com/kubeedge/kubeedge/edge/pkg/metamanager"
	"github.com/kubeedge/kubeedge/edge/pkg/metamanager/client"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/edgecore/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	utilwait "k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/klog/v2"
)

const (
	syncMsgRespTimeout = 1 * time.Minute
	EdgeController     = "edgecontroller"
)

// clusterd is the implementation to manage an edge cluser.
type clusterd struct {
	name                            string
	missionDeployer                 *MissionDeployer
	missionStateRepoter             *MissionStateReporter
	uid                             types.UID
	edgeClusterStatusUpdateInterval time.Duration
	missionStateUpdateInterval      time.Duration
	registrationCompleted           bool
	namespace                       string
	enable                          bool
	metaClient                      client.CoreInterface
	kubeDistro                      string
	kubectlPath                     string
	kubeconfig                      string
}

// Register register clusterd module
func Register(e *v1alpha1.Clusterd) {
	config.InitConfigure(e)
	clusterd, err := newClusterd(e.Enable)
	if err != nil {
		klog.Fatalf("init new clusterd error, %v", err)
	}
	core.Register(clusterd)
}

func (e *clusterd) Name() string {
	return modules.ClusterdModuleName
}

func (e *clusterd) Group() string {
	return modules.ClusterdGroup
}

//Enable indicates whether this module is enabled
func (e *clusterd) Enable() bool {
	return e.enable
}

func (e *clusterd) Start() {
	klog.Info("Starting clusterd...")

	go utilwait.Until(e.syncEdgeClusterStatus, e.edgeClusterStatusUpdateInterval, utilwait.NeverStop)

	klog.Infof("starting sync with cloud")
	e.syncCloud()

	return
}

//newClusterd creates new Clusterd object and initialises it
func newClusterd(enable bool) (*clusterd, error) {
	if !FileExists(config.Config.Kubeconfig) {
		return nil, fmt.Errorf("Could not open kubeconfig file (%s)", config.Config.Kubeconfig)
	}

	if _, exists := DistroToKubectl[config.Config.KubeDistro]; !exists {
		return nil, fmt.Errorf("Invalid kube distribution (%v)", config.Config.KubeDistro)
	}

	basedir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	kubectlPath := filepath.Join(basedir, DistroToKubectl[config.Config.KubeDistro])

	missionDeployer := NewMissionDeployer(&config.Config.Clusterd)

	c := &clusterd{
		name:                            config.Config.Name,
		namespace:                       config.Config.RegisterNamespace,
		missionDeployer:                 missionDeployer,
		enable:                          enable,
		uid:                             types.UID("76246eec-1dc7-4bcf-89b4-686dbc3b4234"),
		edgeClusterStatusUpdateInterval: time.Duration(config.Config.Clusterd.EdgeClusterStatusUpdateInterval) * time.Second,
		metaClient:                      client.New(),
		kubeDistro:                      config.Config.KubeDistro,
		kubeconfig:                      config.Config.Kubeconfig,
		kubectlPath:                     kubectlPath,
	}

	stopChan := make(chan struct{})
	missionStateRepoter := NewMissionStateReporter(&config.Config.Clusterd, c, missionDeployer, stopChan)
	c.missionStateRepoter = missionStateRepoter

	go missionStateRepoter.Run(config.Config.MissionStateWatchWorkers, stopChan)

	return c, nil
}

func (e *clusterd) syncCloud() {
	time.Sleep(10 * time.Second)

	//when starting, send msg to metamanager once to get existing missions
	info := model.NewMessage("").BuildRouter(e.Name(), e.Group(), e.namespace+"/"+model.ResourceTypeMission,
		model.QueryOperation)
	beehiveContext.Send(metamanager.MetaManagerModuleName, *info)
	for {
		select {
		case <-beehiveContext.Done():
			klog.Warning("Clusterd Sync stop")
			return
		default:
		}
		result, err := beehiveContext.Receive(e.Name())
		if err != nil {
			klog.Errorf("failed to sync: %v", err)
			continue
		}

		_, resType, _, err := util.ParseResourceEdge(result.GetResource(), result.GetOperation())
		if err != nil {
			klog.Errorf("failed to parse the Resource: %v", err)
			continue
		}
		op := result.GetOperation()

		var content []byte

		switch result.Content.(type) {
		case []byte:
			content = result.GetContent().([]byte)
		default:
			content, err = json.Marshal(result.Content)
			if err != nil {
				klog.Errorf("marshal message content failed: %v", err)
				continue
			}
		}
		klog.V(4).Infof("result content is %s", result.Content)
		_, resType, resID, err := util.ParseResourceEdge(result.GetResource(), result.GetOperation())
		switch resType {
		case constants.ResourceTypeMission:
			if op == model.ResponseOperation && resID == "" {
				if result.GetSource() != metamanager.MetaManagerModuleName && result.GetSource() != EdgeController {
					klog.Errorf("recevied mission list from unrecognized source : %v", result.GetSource())
					continue
				}
				err := e.handleMissionList(content)
				if err != nil {
					klog.Errorf("handle missionList failed: %v", err)
					continue
				}
			} else {
				err := e.handleMission(op, content)
				if err != nil {
					klog.Errorf("handle mission failed: %v", err)
					continue
				}
			}

		default:
			klog.Errorf("resource type %s is not supported", resType)
			continue
		}
	}
}

func (e *clusterd) handleMissionList(content []byte) (err error) {
	if e.missionDeployer == nil {
		return fmt.Errorf("mission deployer is not initialized.")
	}

	var lists []string
	if err = json.Unmarshal(content, &lists); err != nil {
		return err
	}

	missionList := []*edgeclustersv1.Mission{}
	for _, list := range lists {
		var mission edgeclustersv1.Mission
		err = json.Unmarshal([]byte(list), &mission)
		if err != nil {
			return err
		}
		missionList = append(missionList, &mission)
	}

	return e.missionDeployer.AlignMissionList(missionList)
}

func (e *clusterd) handleMission(op string, content []byte) (err error) {
	if e.missionDeployer == nil {
		return fmt.Errorf("mission deployer is not initialized.")
	}

	var mission edgeclustersv1.Mission
	err = json.Unmarshal(content, &mission)
	if err != nil {
		return err
	}

	switch op {
	case model.InsertOperation:
		err = e.missionDeployer.ApplyMission(&mission)
	case model.UpdateOperation:
		err = e.missionDeployer.ApplyMission(&mission)
	case model.DeleteOperation:
		err = e.missionDeployer.DeleteMission(&mission)
	}
	if err == nil {
		klog.V(3).Infof("%s mission [%s] for cache success.", op, mission.Name)
	}
	return
}
