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
	"strings"

	edgeclustersv1 "github.com/kubeedge/kubeedge/cloud/pkg/apis/edgeclusters/v1"
	"k8s.io/klog/v2"

	"github.com/kubeedge/beehive/pkg/core"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/beehive/pkg/core/model"
	edgeapi "github.com/kubeedge/kubeedge/common/types"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/config"
	clusterdconfig "github.com/kubeedge/kubeedge/edge/pkg/clusterd/config"
	"github.com/kubeedge/kubeedge/edge/pkg/common/message"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
	"github.com/kubeedge/kubeedge/edge/pkg/edgehub"
	v1 "k8s.io/api/core/v1"
)

var initEdgeCluster edgeclustersv1.EdgeCluster

func (e *clusterd) initialEdgeCluster() (*edgeclustersv1.EdgeCluster, error) {
	var ec = &edgeclustersv1.EdgeCluster{}
	var err error

	if err := e.checkclusterdconfig(); err != nil {
		return nil, err
	}

	clusterdconfig := clusterdconfig.Config

	clusterName := clusterdconfig.Name
	if len(clusterName) == 0 {
		clusterName, err = os.Hostname()
		if err != nil {
			klog.Errorf("The cluster name is empty, and couldn't determine hostname: %v", err)
			return nil, err
		}
	}
	ec.Name = clusterName

	ec.Spec.Kubeconfig = clusterdconfig.Kubeconfig
	ec.Spec.KubeDistro = clusterdconfig.KubeDistro

	ec.Labels = map[string]string{
		// Kubernetes built-in labels
		v1.LabelHostname: ec.Name,

		// KubeEdge specific labels
		"role.kubernetes.io/edgecluster": "",
	}

	for k, v := range clusterdconfig.Labels {
		ec.Labels[k] = v
	}

	return ec, nil
}

func (e *clusterd) checkclusterdconfig() error {
	clusterdconfig := clusterdconfig.Config

	if e.TestClusterReady() == false {
		return fmt.Errorf("The cluster is not reacheable.")
	}

	basedir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	mission_crd_file := filepath.Join(basedir, MISSION_CRD_FILE)
	deploy_mission_crd_cmd := fmt.Sprintf("%s apply --kubeconfig=%s -f %s ", e.kubectlPath, clusterdconfig.Kubeconfig, mission_crd_file)
	if _, err := ExecCommandLine(deploy_mission_crd_cmd, COMMAND_TIMEOUT_SEC); err != nil {
		return fmt.Errorf("Failed to deploy mission crd: %v", err)
	}

	edgecluster_crd_file := filepath.Join(basedir, EDGECLUSTER_CRD_FILE)
	deploy_edgecluster_crd_cmd := fmt.Sprintf("%s apply --kubeconfig=%s -f %s ", e.kubectlPath, clusterdconfig.Kubeconfig, edgecluster_crd_file)
	if _, err := ExecCommandLine(deploy_edgecluster_crd_cmd, COMMAND_TIMEOUT_SEC); err != nil {
		return fmt.Errorf("Failed to deploy edgecluster crd: %v", err)
	}

	return nil
}

func (e *clusterd) setInitEdgeCluster(ec *edgeclustersv1.EdgeCluster) {
	initEdgeCluster.Status = *ec.Status.DeepCopy()
}

func (e *clusterd) registerEdgeCluster() error {
	ec, err := e.initialEdgeCluster()
	if err != nil {
		klog.Errorf("Unable to construct edgeclustersv1.EdgeCluster object for edge: %v", err)
		return err
	}

	e.setInitEdgeCluster(ec)

	if !config.Config.RegisterCluster {
		//when register-edgeCluster set to false, do not auto register edgeCluster
		klog.Infof("register-Cluster is set to false")
		e.registrationCompleted = true
		return nil
	}

	resource := fmt.Sprintf("%s/%s/%s", e.namespace, model.ResourceTypeEdgeClusterStatus, ec.Name)

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
	e.registrationCompleted = true

	return nil
}

func (e *clusterd) getEdgeClusterStatusRequest(edgeCluster *edgeclustersv1.EdgeCluster) (*edgeapi.EdgeClusterStatusRequest, error) {
	var edgeClusterStatus = &edgeapi.EdgeClusterStatusRequest{}
	edgeClusterStatus.UID = e.uid
	edgeClusterStatus.Status = *edgeCluster.Status.DeepCopy()
	edgeClusterStatus.Status.Healthy = e.TestClusterReady()

	edgeClusterStatus.Status.EdgeClusters = e.GetEdgeClusterNames()
	edgeClusterStatus.Status.Nodes = e.GetNodeNames()
	edgeClusterStatus.Status.EdgeNodes = e.GetEdgeNodeNames()

	var receivedMissions []string
	var matchededMissions []string
	for k, v := range e.missionDeployer.MissionMatch {
		receivedMissions = append(receivedMissions, k)
		if v == true {
			matchededMissions = append(matchededMissions, k)
		}
	}

	edgeClusterStatus.Status.ReceivedMissions = receivedMissions
	edgeClusterStatus.Status.ActiveMissions = matchededMissions

	klog.V(4).Infof("EdgeCluster Status %#v", edgeClusterStatus)

	return edgeClusterStatus, nil
}

func (e *clusterd) updateEdgeClusterStatus() error {
	edgeClusterStatus, err := e.getEdgeClusterStatusRequest(&initEdgeCluster)
	if err != nil {
		klog.Errorf("Unable to construct api.EdgeClusterStatusRequest object for edge: %v", err)
		return err
	}

	err = e.metaClient.EdgeClusterStatus(e.namespace).Update(e.name, *edgeClusterStatus)
	if err != nil {
		klog.Errorf("update edgeCluster status failed, error: %v", err)
		return err
	}

	return nil
}

func (e *clusterd) UpdateMissionStatus(missionName string, missionStatus map[string]string) error {
	updatedMissionStatus := map[string]string{}
	clusterName := clusterdconfig.Config.Name
	for key, val := range missionStatus {
		if key == LOCAL_EDGE_CLUSTER {
			updatedMissionStatus[clusterName] = val
		} else {
			updatedMissionStatus[clusterName+"/"+key] = val
		}
	}
	msRequest := edgeapi.MissionStatusRequest{
		UID:         e.uid,
		ClusterName: clusterName,
		Status:      updatedMissionStatus,
	}

	err := e.metaClient.MissionStatus(e.namespace).Update(missionName, msRequest)
	if err != nil {
		klog.Errorf("update mission %v status failed, error: %v", missionName, err)
	}

	return err
}

func (e *clusterd) syncEdgeClusterStatus() {
	if !e.registrationCompleted {
		if err := e.registerEdgeCluster(); err != nil {
			klog.Fatalf("Register edgeCluster failed: %v", err)
		}
	}

	if err := e.updateEdgeClusterStatus(); err != nil {
		klog.Errorf("Unable to update edgeCluster status: %v", err)
	}
}

func (e *clusterd) TestClusterReady() bool {
	test_cluster_command := fmt.Sprintf("%s cluster-info --kubeconfig=%s", e.kubectlPath, e.kubeconfig)
	if _, err := ExecCommandLine(test_cluster_command, COMMAND_TIMEOUT_SEC); err != nil {
		klog.Errorf("The cluster is unreachable: %v", err)
		return false
	}

	return true
}

func (e *clusterd) GetEdgeClusterNames() []string {
	return e.GetLocalClusterScopeResourceNames("edgeclusters", "")
}

func (e *clusterd) GetNodeNames() []string {
	return e.GetLocalClusterScopeResourceNames("nodes", "")
}

func (e *clusterd) GetEdgeNodeNames() []string {
	return e.GetLocalClusterScopeResourceNames("nodes", "node-role.kubernetes.io/edge=")
}

func (e *clusterd) GetLocalClusterScopeResourceNames(resType string, label string) []string {
	labelOption := ""
	if len(label) > 0 {
		labelOption = "-l " + label
	}
	get_resource_cmd := fmt.Sprintf(" %s get %s -o json %s --kubeconfig=%s | jq -r '.items[] | [.metadata.name] | @tsv' ", e.kubectlPath, resType, labelOption, e.kubeconfig)
	output, err := ExecCommandLine(get_resource_cmd, COMMAND_TIMEOUT_SEC)
	if err != nil {
		klog.Errorf("Failed to get %v: %v", err)
		return []string{"error"}
	}

	names := []string{}
	for _, o := range strings.Split(output, "\n") {
		name := strings.TrimSpace(o)
		if len(name) > 0 {
			names = append(names, name)
		}
	}

	return names
}
