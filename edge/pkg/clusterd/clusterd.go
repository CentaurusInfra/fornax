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
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	"github.com/kubeedge/beehive/pkg/common/util"
	"github.com/kubeedge/beehive/pkg/core"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/beehive/pkg/core/model"
	"github.com/kubeedge/kubeedge/common/constants"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/config"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
	"github.com/kubeedge/kubeedge/edge/pkg/metamanager"
	"github.com/kubeedge/kubeedge/edge/pkg/metamanager/client"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/edgecore/v1alpha1"
)

const (
	syncMsgRespTimeout = 1 * time.Minute
	EdgeController     = "edgecontroller"
)

// clusterd is the implementation to manage an edge cluster.
type clusterd struct {
	missionDeployer          *MissionDeployer
	missionStateRepoter      *MissionStateReporter
	edgeClusterStateReporter *EdgeClusterStateReporter
	uid                      types.UID
	namespace                string
	enable                   bool
	metaClient               client.CoreInterface
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

	klog.Infof("starting sync with cloud")
	e.syncCloud()
}

//newClusterd creates new Clusterd object and initialises it
func newClusterd(enable bool) (*clusterd, error) {
	missionDeployer := NewMissionDeployer()

	c := &clusterd{
		namespace:       config.Config.RegisterNamespace,
		missionDeployer: missionDeployer,
		enable:          enable,
		uid:             types.UID("76246eec-1dc7-4bcf-89b4-686dbc3b4234"),
		metaClient:      client.New(),
	}

	c.missionStateRepoter = NewMissionStateReporter(c, missionDeployer)
	c.edgeClusterStateReporter = NewEdgeClusterStateReporter(c, missionDeployer)

	go c.missionStateRepoter.Run()
	go c.edgeClusterStateReporter.Run()

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

		_, resType, _, err := util.ParseResourceEdgeCluster(result.GetResource(), result.GetOperation())
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
		_, resType, resID, err := util.ParseResourceEdgeCluster(result.GetResource(), result.GetOperation())
		if err != nil {
			klog.Errorf("failed in edge resource parsing: %v", err)
			continue
		}
		switch resType {
		case constants.ResourceTypeMission:
			if op == model.ResponseOperation && resID == "" {
				if result.GetSource() != metamanager.MetaManagerModuleName && result.GetSource() != EdgeController {
					klog.Errorf("recevied mission list from unrecognized source : %v", result.GetSource())
					continue
				}
				err := e.missionDeployer.UnmarshalAndHandleMissionStringList(content)
				if err != nil {
					klog.Errorf("handle missionList failed: %v", err)
					continue
				}
			} else {
				err := e.missionDeployer.UnmarshalAndHandleMission(op, content)
				if err != nil {
					klog.Errorf("handle mission failed: %v", err)
					continue
				}
			}
		case constants.ResourceTypeMissionList:
			if result.GetSource() != metamanager.MetaManagerModuleName && result.GetSource() != EdgeController {
				klog.Errorf("recevied missionlist from unrecognized source : %v", result.GetSource())
				continue
			}

			err := e.missionDeployer.UnmarshalAndHandleMissionObjectList(content)
			if err != nil {
				klog.Errorf("handle missionList failed: %v", err)
				continue
			}

		default:
			klog.Errorf("resource type %s is not supported", resType)
			continue
		}
	}
}
