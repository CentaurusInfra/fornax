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
	"context"
	"encoding/json"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strings"

	edgeclustersv1 "github.com/kubeedge/kubeedge/cloud/pkg/apis/edgeclusters/v1"
	crdClientset "github.com/kubeedge/kubeedge/cloud/pkg/client/clientset/versioned"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/edgecore/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/klog/v2"
)

const (
	COMMAND_TIMEOUT_SEC  = 10
	MISSION_CRD_FILE     = "mission_v1.yaml"
	EDGECLUSTER_CRD_FILE = "edgecluster_v1.yaml"
	LOCAL_EDGE_CLUSTER   = "local_edge_cluster"
	STATUS_NO_MATCH      = "not match"
)

var DistroToKubectl = map[string]string{
	"arktos": "kubectl/arktos/kubectl",
	"k8s":    "kubectl/vanilla/kubectl",
}

type MissionDeployer struct {
	ClusterName    string
	ClusterLabels  map[string]string
	KubeDistro     string
	KubeconfigFile string
	KubectlCli     string
	MissionMatch   map[string]bool
	CrdClient      *crdClientset.Clientset
}

//NewMissionDeployer creates new mission deployer object
func NewMissionDeployer(clusterdConfig *v1alpha1.Clusterd) *MissionDeployer {

	// No need to check the clusterdConfig, as it was checked during the clusterd initialization
	basedir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", clusterdConfig.Kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to build config, err: %v", err)
	}

	crdKubeConfig := rest.CopyConfig(kubeConfig)
	crdKubeConfig.ContentType = runtime.ContentTypeJSON

	return &MissionDeployer{
		ClusterName:    clusterdConfig.Name,
		ClusterLabels:  clusterdConfig.Labels,
		KubeDistro:     clusterdConfig.KubeDistro,
		KubeconfigFile: clusterdConfig.Kubeconfig,
		KubectlCli:     filepath.Join(basedir, DistroToKubectl[clusterdConfig.KubeDistro]),
		MissionMatch:   map[string]bool{},
		CrdClient:      crdClientset.NewForConfigOrDie(crdKubeConfig),
	}
}

func (m *MissionDeployer) ApplyMission(mission *edgeclustersv1.Mission) error {
	m.MissionMatch[mission.Name] = m.isMatchingMission(mission)

	needUpdateStatus := m.checkNeedUpdateStatus(mission)

	missionYaml, err := buildMissionYaml(mission)
	if err != nil {
		// log the error and move on to apply the mission content
		klog.Errorf("Error in applying mission CRD: %v. Moving on.", err)
	} else {
		deploy_mission_cmd := fmt.Sprintf("printf \"%s\" | %s apply --kubeconfig=%s -f - ", missionYaml, m.KubectlCli, m.KubeconfigFile)
		_, err := ExecCommandLine(deploy_mission_cmd, COMMAND_TIMEOUT_SEC)
		if err != nil {
			klog.Errorf("Failed to apply mission %v: %v", mission.Name, err)
		} else {
			klog.V(3).Infof("Mission %v is saved.", mission.Name)
		}
	}

	if m.isMatchingMission(mission) == false {
		if needUpdateStatus {
			m.UpdateMissionLocalStatus(mission.Name, STATUS_NO_MATCH)
		}
		klog.V(3).Infof("Mission %v does not match this cluster, skip the content applying", mission.Name)
		return nil
	}

	if strings.TrimSpace(mission.Spec.Content) != "" {
		deploy_content_cmd := fmt.Sprintf("printf \"%s\" | %s apply --kubeconfig=%s -f - ", mission.Spec.Content, m.KubectlCli, m.KubeconfigFile)
		_, err = ExecCommandLine(deploy_content_cmd, COMMAND_TIMEOUT_SEC)
		if err != nil {
			klog.Errorf("Failed to apply the content of mission %v: %v", mission.Name, err)
		} else {
			klog.V(2).Infof("The content of mission %v applied successfully ", mission.Name)
		}
	}

	if needUpdateStatus {
		m.StatusUpdate(mission, false)
	}

	return nil
}

func (m *MissionDeployer) DeleteMission(mission *edgeclustersv1.Mission) error {
	delete(m.MissionMatch, mission.Name)
	if m.isMatchingMission(mission) == false {
		klog.V(4).Infof("Mission %v does not match this cluster", mission.Name)
	} else {
		if strings.TrimSpace(mission.Spec.Content) != "" {
			delete_content_cmd := fmt.Sprintf("printf \"%s\" | %s delete --kubeconfig=%s -f - ", mission.Spec.Content, m.KubectlCli, m.KubeconfigFile)
			_, err := ExecCommandLine(delete_content_cmd, COMMAND_TIMEOUT_SEC)
			if err != nil {
				klog.Errorf("Failed to revert the content of mission %v: %v", mission.Name, err)
			} else {
				klog.Errorf("The content of mission %v is reverted.", mission.Name)
			}
		}
	}

	delete_mission_cmd := fmt.Sprintf("%s delete mission %s --kubeconfig=%s", m.KubectlCli, mission.Name, m.KubeconfigFile)
	if _, err := ExecCommandLine(delete_mission_cmd, COMMAND_TIMEOUT_SEC); err != nil {
		return fmt.Errorf("Failed to delete mission %v: %v", mission.Name, err)
	}

	klog.Infof("Mission %v deleted successfully ", mission.Name)

	return nil
}

func (m *MissionDeployer) DeleteMissionByName(name string) error {
	get_mission_cmd := fmt.Sprintf("%s get mission %s --kubeconfig=%s -o json ", m.KubectlCli, name, m.KubeconfigFile)
	output, err := ExecCommandLine(get_mission_cmd, COMMAND_TIMEOUT_SEC)
	if err != nil {
		return fmt.Errorf("Failed to get mission %v: %v", name, err)
	}

	var mission edgeclustersv1.Mission
	err = json.Unmarshal([]byte(output), &mission)
	if err != nil {
		return err
	}

	return m.DeleteMission(&mission)
}

func (m *MissionDeployer) GetLocalMissionNames() ([]string, error) {
	get_mission_cmd := fmt.Sprintf(" %s get missions -o json --kubeconfig=%s | jq -r '.items[] | [.metadata.name] | @tsv' ", m.KubectlCli, m.KubeconfigFile)
	output, err := ExecCommandLine(get_mission_cmd, COMMAND_TIMEOUT_SEC)
	if err != nil {
		return nil, fmt.Errorf("Failed to get missions: %v", err)
	}

	names := []string{}
	for _, o := range strings.Split(output, "\n") {
		name := strings.TrimSpace(o)
		if len(name) > 0 {
			names = append(names, name)
		}
	}

	return names, nil
}

func (m *MissionDeployer) isMatchingMission(mission *edgeclustersv1.Mission) bool {
	// if the placement field is empty, it matches all the edge clusters
	if len(mission.Spec.Placement.Clusters) == 0 && len(mission.Spec.Placement.MatchLabels) == 0 {
		return true
	}

	for _, matchingCluster := range mission.Spec.Placement.Clusters {
		if m.ClusterName == matchingCluster.Name {
			return true
		}
	}

	// TODO: use k8s Labels operator to match
	if len(mission.Spec.Placement.MatchLabels) == 0 {
		return false
	}

	for k, v := range mission.Spec.Placement.MatchLabels {
		if val, ok := m.ClusterLabels[k]; ok && val == v {
			return true
		}
	}

	return false
}

func (m *MissionDeployer) AlignMissionList(missionList []*edgeclustersv1.Mission) error {
	missionMap := map[string]bool{}
	var errs []error
	for _, mi := range missionList {
		missionMap[mi.Name] = true
		if err := m.ApplyMission(mi); err != nil {
			// Try to apply as many missions as possible, so move on after hitting error
			errs = append(errs, fmt.Errorf("Error when applying mission %s: %v", mi.Name, err))
		}
	}

	localMissions, err := m.GetLocalMissionNames()
	if err != nil {
		errs = append(errs, fmt.Errorf("Error when get local missions: %v", err))
		return fmt.Errorf("Hit the errors in mission align: %v", errs)
	}

	for _, mi := range localMissions {
		if _, exists := missionMap[mi]; !exists {
			if err := m.DeleteMissionByName(mi); err != nil {
				// Try to remove as many missions as possible, so move on after hitting error
				errs = append(errs, fmt.Errorf("Error when deleting mission %s: %v", mi, err))
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("Hit the errors in mission align: %v", errs)
}

// create a yaml to use by "kubectl apply" command
func buildMissionYaml(input *edgeclustersv1.Mission) (string, error) {
	yaml_part1_template := `apiVersion: edgeclusters.kubeedge.io/v1
kind: Mission
metadata:
  name: %s
spec:
  %s`
	specStr, err := yaml.Marshal(input.Spec)
	if err != nil {
		return "", err
	}

	output := fmt.Sprintf(yaml_part1_template, input.Name, strings.ReplaceAll(string(specStr), "\n", "\n  "))
	return output, nil
}

func (m *MissionDeployer) UpdateMissionLocalStatus(name string, statusInfo string) error {
	statusInfo = strings.TrimSpace(statusInfo)
	mission, err := m.CrdClient.EdgeclustersV1().Missions().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if mission.Status == nil {
		mission.Status = map[string]string{}
	} else {
		v, ok := mission.Status[LOCAL_EDGE_CLUSTER]
		if ok && v == statusInfo {
			klog.V(3).Infof("The status of local cluster is not changed.")
			return nil
		}
	}

	mission.Status[LOCAL_EDGE_CLUSTER] = statusInfo

	_, err = m.CrdClient.EdgeclustersV1().Missions().UpdateStatus(context.Background(), mission, metav1.UpdateOptions{})

	return err
}

// this is a hacky way for PoC only
func analyzeMissionContent(content string) (kind string, name string, namespace string) {
	for _, line := range strings.Split(content, "\n") {
		words := strings.Split(strings.TrimSpace(line), " ")
		if strings.Contains(line, "kind:") && kind == "" {
			kind = words[len(words)-1]
			continue
		}
		if strings.Contains(line, "name:") && name == "" {
			name = words[len(words)-1]
			continue
		}
		if strings.Contains(line, "namespace:") && namespace == "" {
			namespace = words[len(words)-1]
			continue
		}
	}
	return
}

func (m *MissionDeployer) GetStatusCheckCommand(mission *edgeclustersv1.Mission) string {

	command := strings.TrimSpace(mission.Spec.StatusCheck.Command)
	if command != "" {
		return strings.ReplaceAll(command, "${kubectl}", m.KubectlCli)
	}

	kind, name, namespace := analyzeMissionContent(mission.Spec.Content)

	command = fmt.Sprintf("%v get %v %v -n \"%v\" --kubeconfig %v --no-headers", m.KubectlCli, kind, name, namespace, m.KubeconfigFile)

	klog.V(3).Infof("the status check command is %v ", command)

	return command
}

func (m *MissionDeployer) StatusUpdate(mission *edgeclustersv1.Mission, needCheckMatch bool) {
	if needCheckMatch && !m.isMatchingMission(mission) {
		if err := m.UpdateMissionLocalStatus(mission.Name, STATUS_NO_MATCH); err != nil {
			klog.Errorf("Error when updating the mission %v status: %v", mission.Name, err)
		}
		return
	}

	status_command := m.GetStatusCheckCommand(mission)
	output, err := ExecCommandLine(status_command, COMMAND_TIMEOUT_SEC)
	if err != nil {
		klog.Errorf("Error when checking the mission %v status: %v, output %v", mission.Name, err, output)
	}

	err = m.UpdateMissionLocalStatus(mission.Name, output)
	if err != nil {
		klog.Errorf("Error when updating the mission %v status: %v", mission.Name, err)
	}
}

// We should only update the status if there is change in the mission Spec.
// NO need to update the status if the change is in the Mission status.
// Otherwise, the system will be drained, as the clusterd will be trapped
// in getting update event which is caused by its own status update action and making another status update action.
func (m *MissionDeployer) checkNeedUpdateStatus(mission *edgeclustersv1.Mission) bool {
	existingMission, err := m.CrdClient.EdgeclustersV1().Missions().Get(context.Background(), mission.Name, metav1.GetOptions{})
	if err != nil {
		// either not found or some other error, let's check the status
		klog.Infof("Error in gettting mission %v : %v ", mission.Name, err)
		return true
	}

	if !TrueEqual(existingMission.Spec, mission.Spec) {
		klog.Infof("Mission %v Spec has changed. exist(%#v) new (%#v)", mission.Name, existingMission.Spec, mission.Spec)
		return true
	}

	return false

}

// for some reason we still need to find out, the same mission spec objects may be no longer deep-equal.
// For instance, a null array turns into an empty array. This function aims to detect two spec objects are truly equal.
func TrueEqual(a edgeclustersv1.MissionSpec, b edgeclustersv1.MissionSpec) bool {
	if a.Content != b.Content {
		return false
	}

	if a.StatusCheck.Command != b.StatusCheck.Command {
		return false
	}

	if !EqualMaps(a.Placement.MatchLabels, b.Placement.MatchLabels) {
		return false
	}

	if !EqualArray(a.Placement.Clusters, b.Placement.Clusters) {
		return false
	}

	return true
}
