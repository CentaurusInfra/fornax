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
*/

package clusterd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/kubeedge/beehive/pkg/core"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/beehive/pkg/core/model"
	edgeclustersv1 "github.com/kubeedge/kubeedge/cloud/pkg/apis/edgeclusters/v1"
	edgeapi "github.com/kubeedge/kubeedge/common/types"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/config"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/helper"
	"github.com/kubeedge/kubeedge/edge/pkg/common/message"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
	"github.com/kubeedge/kubeedge/edge/pkg/edgehub"
)

const (
	crdFolder       = "crds"
	HealthyStatus   = "healthy"
	UnhealthyStatus = "unhealthy"
)

var initEdgeCluster edgeclustersv1.EdgeCluster

type EdgeClusterStateReporter struct {
	clusterd                       *clusterd
	edgeClusterStateUpdateInterval time.Duration
	missionDeployer                *MissionDeployer
	registrationCompleted          bool
}

func NewEdgeClusterStateReporter(c *clusterd, md *MissionDeployer) *EdgeClusterStateReporter {
	return &EdgeClusterStateReporter{
		clusterd:                       c,
		missionDeployer:                md,
		edgeClusterStateUpdateInterval: time.Duration(config.Config.EdgeClusterStateUpdateInterval) * time.Second,
		registrationCompleted:          false,
	}
}

func (esr *EdgeClusterStateReporter) initialEdgeCluster() (*edgeclustersv1.EdgeCluster, error) {
	var ec = &edgeclustersv1.EdgeCluster{}

	if err := esr.prepareCluster(); err != nil {
		return nil, err
	}

	ec.Name = config.Config.Name
	ec.Spec.Kubeconfig = config.Config.Kubeconfig
	ec.Spec.KubeDistro = config.Config.KubeDistro
	ec.Labels = config.Config.Labels

	return ec, nil
}

func (esr *EdgeClusterStateReporter) prepareCluster() error {
	basedir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	crdFilePath := filepath.Join(basedir, crdFolder)
	deployCrdCmd := fmt.Sprintf("%s apply --kubeconfig=%s -f %s ", config.Config.KubectlCli, config.Config.Kubeconfig, crdFilePath)
	if _, err := helper.ExecCommandToCluster(deployCrdCmd); err != nil {
		return fmt.Errorf("Failed to deploy crd files: %v", err)
	}

	return nil
}

func (esr *EdgeClusterStateReporter) setInitEdgeClusterState(ec *edgeclustersv1.EdgeCluster) {
	initEdgeCluster.State = *ec.State.DeepCopy()
}

func (esr *EdgeClusterStateReporter) registerEdgeCluster() error {
	ec, err := esr.initialEdgeCluster()
	if err != nil {
		klog.Errorf("Unable to construct edgeclustersv1.EdgeCluster object for edge: %v", err)
		return err
	}

	esr.setInitEdgeClusterState(ec)

	if !config.Config.RegisterCluster {
		//when register-edgeCluster set to false, do not auto register edgeCluster
		klog.Infof("register-Cluster is set to false")
		esr.registrationCompleted = true
		return nil
	}

	resource := fmt.Sprintf("%s/%s/%s", esr.clusterd.namespace, model.ResourceTypeEdgeClusterState, ec.Name)
	klog.Infof("Attempting to register edgeCluster (%s), %s", ec.Name, resource)
	edgeClusterInfoMsg := message.BuildMsg(modules.MetaGroup, "", modules.ClusterdModuleName, resource, model.InsertOperation, ec)

	var res model.Message
	if _, ok := core.GetModules()[edgehub.ModuleNameEdgeHub]; ok {
		res, err = beehiveContext.SendSync(edgehub.ModuleNameEdgeHub, *edgeClusterInfoMsg, syncMsgRespTimeout)
	} else {
		res, err = beehiveContext.SendSync(EdgeController, *edgeClusterInfoMsg, syncMsgRespTimeout)
	}

	if err != nil || res.Content != "OK" {
		klog.Errorf("register edgeCluster failed, error: %v", err)
		if res.Content != "OK" {
			klog.Errorf("response from cloud core: %v", res.Content)
		}
		return err
	}

	klog.Infof("Successfully registered edgeCluster %s", ec.Name)
	esr.registrationCompleted = true

	return nil
}

func (esr *EdgeClusterStateReporter) getEdgeClusterStateRequest(edgeCluster *edgeclustersv1.EdgeCluster) (*edgeapi.EdgeClusterStateRequest, error) {
	var edgeClusterState = &edgeapi.EdgeClusterStateRequest{}
	edgeClusterState.UID = esr.clusterd.uid
	edgeClusterState.State = *edgeCluster.State.DeepCopy()

	edgeClusterState.State.HealthStatus = GetLocalClusterStatus()

	if edgeClusterState.State.HealthStatus == HealthyStatus {
		edgeClusterState.State.SubEdgeClusterStates = GetSubEdgeClusterStates()
		edgeClusterState.State.Nodes = helper.GetLocalClusterScopeResourceNames("nodes", "")
		edgeClusterState.State.EdgeNodes = helper.GetLocalClusterScopeResourceNames("nodes", "node-role.kubernetes.io/edge")

		var receivedMissions []string
		var matchededMissions []string
		for k, v := range esr.missionDeployer.MissionMatch {
			receivedMissions = append(receivedMissions, k)
			if v {
				matchededMissions = append(matchededMissions, k)
			}
		}

		edgeClusterState.State.ReceivedMissions = receivedMissions
		edgeClusterState.State.ActiveMissions = matchededMissions
	} else {
		edgeClusterState.State.SubEdgeClusterStates = map[string]string{}
		edgeClusterState.State.Nodes = []string{}
		edgeClusterState.State.EdgeNodes = []string{}
		edgeClusterState.State.ReceivedMissions = []string{}
		edgeClusterState.State.ActiveMissions = []string{}
	}

	klog.V(4).Infof("EdgeCluster Status %#v", edgeClusterState)

	return edgeClusterState, nil
}

func (esr *EdgeClusterStateReporter) updateEdgeClusterState() error {
	edgeClusterState, err := esr.getEdgeClusterStateRequest(&initEdgeCluster)
	if err != nil {
		klog.Errorf("Unable to construct api.EdgeClusterStateRequest object for edge: %v", err)
		return err
	}

	err = esr.clusterd.metaClient.EdgeClusterState(esr.clusterd.namespace).Update(config.Config.Name, *edgeClusterState)
	if err != nil {
		klog.Errorf("update edgeCluster status failed, error: %v", err)
		return err
	}

	return nil
}

func (esr *EdgeClusterStateReporter) syncEdgeClusterState() {
	if !esr.registrationCompleted {
		if err := esr.registerEdgeCluster(); err != nil {
			klog.Errorf("Register edgeCluster failed: %v", err)
		}
	}

	if err := esr.updateEdgeClusterState(); err != nil {
		klog.Errorf("Unable to update edgeCluster status: %v", err)
	}
}

func (esr *EdgeClusterStateReporter) Run() {
	klog.Infof("Starting edgecluster state reporter.")
	defer klog.Infof("Shutting down edgecluster state reporter")

	go utilwait.Until(esr.syncEdgeClusterState, esr.edgeClusterStateUpdateInterval, utilwait.NeverStop)
}

func GetSubEdgeClusterStates() map[string]string {
	aggregatedState := map[string]string{}

	getEcStatesCmd := fmt.Sprintf(" %s get edgeclusters -o json --kubeconfig=%s | jq -r '.items[] | {(.metadata.name): .state}' ", config.Config.KubectlCli, config.Config.Kubeconfig)
	output, err := helper.ExecCommandToCluster(getEcStatesCmd)
	if err != nil {
		klog.Errorf("Failed to get edgecluster states: %v", err)
		return aggregatedState
	}

	if strings.TrimSpace(output) == "" {
		klog.V(4).Infof("There is no edge clusters.")
		return aggregatedState
	}

	var ecState map[string]edgeclustersv1.EdgeClusterState

	if err := json.Unmarshal([]byte(output), &ecState); err != nil {
		klog.Errorf("Error in unmarshall edgecluster state json: (%s), error: %v", output, err)
		return aggregatedState
	}

	for edgeCluster, ecState := range ecState {
		aggregatedState[edgeCluster] = ecState.HealthStatus
		for subCluster, state := range ecState.SubEdgeClusterStates {
			aggregatedState[edgeCluster+"/"+subCluster] = state
		}
	}

	return aggregatedState
}

func GetLocalClusterStatus() string {
	clusterHealthy := helper.TestClusterReady()

	var status string
	if clusterHealthy {
		status = HealthyStatus
	} else {
		status = UnhealthyStatus
	}

	return status
}
