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
1. package edgecluster got some functions from "k8s.io/kubernetes/pkg/kubelet/kubelet.go"
and made some variant
*/

package edgecluster

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kubeedge/beehive/pkg/common/util"
	"github.com/kubeedge/beehive/pkg/core"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/beehive/pkg/core/model"
	edgeclustersv1 "github.com/kubeedge/kubeedge/cloud/pkg/apis/edgeclusters/v1"
	"github.com/kubeedge/kubeedge/common/constants"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
	"github.com/kubeedge/kubeedge/edge/pkg/edgecluster/config"
	"github.com/kubeedge/kubeedge/edge/pkg/metamanager"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/edgecore/v1alpha1"
	utilwait "k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/klog/v2"
)

const (
	syncMsgRespTimeout = 1 * time.Minute
	EdgeController = "edgecontroller"
)

// edgeCluster is the main edgeCluster implementation.
type edgeCluster struct {
	name                  string
	missionManager        *MissionManager
	statusUpdateInterval  time.Duration
	registrationCompleted bool
	namespace             string
	enable                bool
}

// Register register edgeCluster
func Register(e *v1alpha1.EdgeCluster) {
	config.InitConfigure(e)
	edgeCluster, err := newEdgeCluster(e.Enable)
	if err != nil {
		klog.Errorf("init new edgeCluster error, %v", err)
		os.Exit(1)
		return
	}
	core.Register(edgeCluster)
}

func (e *edgeCluster) Name() string {
	return modules.EdgeClusterModuleName
}

func (e *edgeCluster) Group() string {
	return modules.EdgeClusterGroup
}

//Enable indicates whether this module is enabled
func (e *edgeCluster) Enable() bool {
	return e.enable
}

func (e *edgeCluster) Start() {
	klog.Info("Starting edgeCluster...")

	go utilwait.Until(e.syncEdgeClusterStatus, e.statusUpdateInterval, utilwait.NeverStop)

	klog.Infof("starting sync with cloud")
	e.syncCloud()

	return
}

//newEdgeCluster creates new edgeCluster object and initialises it
func newEdgeCluster(enable bool) (*edgeCluster, error) {
	missionManager := NewMissionManager(&config.Config.EdgeCluster)

	ec := &edgeCluster{
		name:                 config.Config.Name,
		namespace:            config.Config.RegisterNamespace,
		missionManager:       missionManager,
		enable:               enable,
		statusUpdateInterval: time.Duration(config.Config.EdgeCluster.StatusUpdateInterval) * time.Second,
	}
	return ec, nil
}

func (e *edgeCluster) syncCloud() {
	time.Sleep(10 * time.Second)

	//when starting, send msg to metamanager once to get existing missions
	info := model.NewMessage("").BuildRouter(e.Name(), e.Group(), e.namespace+"/"+model.ResourceTypeMission,
		model.QueryOperation)
	beehiveContext.Send(metamanager.MetaManagerModuleName, *info)
	for {
		select {
		case <-beehiveContext.Done():
			klog.Warning("EdgeCluster Sync stop")
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

func (e *edgeCluster) handleMissionList(content []byte) (err error) {
	if e.missionManager == nil {
		return fmt.Errorf("mission manager is not initialized.")
	}

	var lists []string
	if err = json.Unmarshal([]byte(content), &lists); err != nil {
		return err
	}

	var missionList []*edgeclustersv1.Mission

	for _, list := range lists {
		var mission edgeclustersv1.Mission
		err = json.Unmarshal([]byte(list), &mission)
		if err != nil {
			return err
		}
		missionList = append(missionList, &mission)
	}

	return e.missionManager.AlignMissionList(missionList)
}

func (e *edgeCluster) handleMission(op string, content []byte) (err error) {
	if e.missionManager == nil {
		return fmt.Errorf("mission manager is not initialized.")
	}

	var mission edgeclustersv1.Mission
	err = json.Unmarshal(content, &mission)
	if err != nil {
		return err
	}

	switch op {
	case model.InsertOperation:
		err = e.missionManager.ApplyMission(&mission)
	case model.UpdateOperation:
		err = e.missionManager.ApplyMission(&mission)
	case model.DeleteOperation:
		err = e.missionManager.DeleteMission(&mission)
	}
	if err == nil {
		klog.Infof("%s mission [%s] for cache success.", op, mission.Name)
	}
	return
}
