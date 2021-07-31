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
	"fmt"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/kubeedge/beehive/pkg/core"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/beehive/pkg/core/model"
	edgeclustersv1 "github.com/kubeedge/kubeedge/cloud/pkg/apis/edgeclusters/v1"
	edgeapi "github.com/kubeedge/kubeedge/common/types"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/config"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/helper"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/util"
	"github.com/kubeedge/kubeedge/edge/pkg/common/message"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
	"github.com/kubeedge/kubeedge/edge/pkg/edgehub"
)

const (
	MissionCrdFile     = "mission_v1.yaml"
	EdgeClusterCrdFile = "edgecluster_v1.yaml"
)

var initEdgeCluster edgeclustersv1.EdgeCluster

type EdgeClusterStatusReporter struct {
	clusterd                        *clusterd
	edgeClusterStatusUpdateInterval time.Duration
	missionDeployer                 *MissionDeployer
	registrationCompleted           bool
}

func NewEdgeClusterStatusReporter(c *clusterd, md *MissionDeployer) *EdgeClusterStatusReporter {
	return &EdgeClusterStatusReporter{
		clusterd:                        c,
		missionDeployer:                 md,
		edgeClusterStatusUpdateInterval: time.Duration(config.Config.EdgeClusterStatusUpdateInterval) * time.Second,
		registrationCompleted:           false,
	}
}

func (esr *EdgeClusterStatusReporter) initialEdgeCluster() (*edgeclustersv1.EdgeCluster, error) {
	var ec = &edgeclustersv1.EdgeCluster{}

	if err := esr.prepareCluster(); err != nil {
		return nil, err
	}

	ec.Name = config.Config.Name
	ec.Spec.Kubeconfig = config.Config.Kubeconfig
	ec.Spec.KubeDistro = config.Config.KubeDistro

	ec.Labels = map[string]string{
		// Kubernetes built-in labels
		v1.LabelHostname: ec.Name,

		// KubeEdge specific labels
		"role.kubernetes.io/edgecluster": "",
	}

	for k, v := range config.Config.Labels {
		ec.Labels[k] = v
	}

	return ec, nil
}

func (esr *EdgeClusterStatusReporter) prepareCluster() error {
	if !helper.TestClusterReady() {
		return fmt.Errorf("the cluster is not reacheable")
	}

	basedir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	missionCrdFilePath := filepath.Join(basedir, MissionCrdFile)
	deployMissioCrdCmd := fmt.Sprintf("%s apply --kubeconfig=%s -f %s ", config.Config.KubectlCli, config.Config.Kubeconfig, missionCrdFilePath)
	if _, err := util.ExecCommandLine(deployMissioCrdCmd); err != nil {
		return fmt.Errorf("Failed to deploy mission crd: %v", err)
	}

	edgeClusterCrdFilePath := filepath.Join(basedir, EdgeClusterCrdFile)
	deployEdgeClusterCrdCmd := fmt.Sprintf("%s apply --kubeconfig=%s -f %s ", config.Config.KubectlCli, config.Config.Kubeconfig, edgeClusterCrdFilePath)
	if _, err := util.ExecCommandLine(deployEdgeClusterCrdCmd); err != nil {
		return fmt.Errorf("Failed to deploy edgecluster crd: %v", err)
	}

	return nil
}

func (esr *EdgeClusterStatusReporter) setInitEdgeClusterStatus(ec *edgeclustersv1.EdgeCluster) {
	initEdgeCluster.Status = *ec.Status.DeepCopy()
}

func (esr *EdgeClusterStatusReporter) registerEdgeCluster() error {
	ec, err := esr.initialEdgeCluster()
	if err != nil {
		klog.Errorf("Unable to construct edgeclustersv1.EdgeCluster object for edge: %v", err)
		return err
	}

	esr.setInitEdgeClusterStatus(ec)

	if !config.Config.RegisterCluster {
		//when register-edgeCluster set to false, do not auto register edgeCluster
		klog.Infof("register-Cluster is set to false")
		esr.registrationCompleted = true
		return nil
	}

	resource := fmt.Sprintf("%s/%s/%s", esr.clusterd.namespace, model.ResourceTypeEdgeClusterStatus, ec.Name)
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

func (esr *EdgeClusterStatusReporter) getEdgeClusterStatusRequest(edgeCluster *edgeclustersv1.EdgeCluster) (*edgeapi.EdgeClusterStatusRequest, error) {
	var edgeClusterStatus = &edgeapi.EdgeClusterStatusRequest{}
	edgeClusterStatus.UID = esr.clusterd.uid
	edgeClusterStatus.Status = *edgeCluster.Status.DeepCopy()
	edgeClusterStatus.Status.Healthy = helper.TestClusterReady()

	edgeClusterStatus.Status.EdgeClusters = helper.GetLocalClusterScopeResourceNames("edgeclusters", "")
	edgeClusterStatus.Status.Nodes = helper.GetLocalClusterScopeResourceNames("nodes", "")
	edgeClusterStatus.Status.EdgeNodes = helper.GetLocalClusterScopeResourceNames("nodes", "node-role.kubernetes.io/edge")

	var receivedMissions []string
	var matchededMissions []string
	for k, v := range esr.missionDeployer.MissionMatch {
		receivedMissions = append(receivedMissions, k)
		if v {
			matchededMissions = append(matchededMissions, k)
		}
	}

	edgeClusterStatus.Status.ReceivedMissions = receivedMissions
	edgeClusterStatus.Status.ActiveMissions = matchededMissions

	klog.V(4).Infof("EdgeCluster Status %#v", edgeClusterStatus)

	return edgeClusterStatus, nil
}

func (esr *EdgeClusterStatusReporter) updateEdgeClusterStatus() error {
	edgeClusterStatus, err := esr.getEdgeClusterStatusRequest(&initEdgeCluster)
	if err != nil {
		klog.Errorf("Unable to construct api.EdgeClusterStatusRequest object for edge: %v", err)
		return err
	}

	err = esr.clusterd.metaClient.EdgeClusterStatus(esr.clusterd.namespace).Update(config.Config.Name, *edgeClusterStatus)
	if err != nil {
		klog.Errorf("update edgeCluster status failed, error: %v", err)
		return err
	}

	return nil
}

func (esr *EdgeClusterStatusReporter) syncEdgeClusterStatus() {
	if !esr.registrationCompleted {
		if err := esr.registerEdgeCluster(); err != nil {
			klog.Errorf("Register edgeCluster failed: %v", err)
		}
	}

	if err := esr.updateEdgeClusterStatus(); err != nil {
		klog.Errorf("Unable to update edgeCluster status: %v", err)
	}
}

func (esr *EdgeClusterStatusReporter) Run() {
	klog.Infof("Starting edgecluster state reporter.")
	defer klog.Infof("Shutting down edgecluster state reporter")

	go utilwait.Until(esr.syncEdgeClusterStatus, esr.edgeClusterStatusUpdateInterval, utilwait.NeverStop)
}
