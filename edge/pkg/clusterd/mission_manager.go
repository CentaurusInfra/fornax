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
	yaml "gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strings"

	edgeclustersv1 "github.com/kubeedge/kubeedge/cloud/pkg/apis/edgeclusters/v1"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/edgecore/v1alpha1"

	"k8s.io/klog/v2"
)

const (
	COMMAND_TIMEOUT_SEC  = 10
	MISSION_CRD_FILE     = "mission_v1.yaml"
	EDGECLUSTER_CRD_FILE = "edgecluster_v1.yaml"
)

var DistroToKubectl = map[string]string{
	"arktos": "kubectl/arktos/kubectl",
	"k8s":    "kubectl/vanilla/kubectl",
}

type MissionManager struct {
	ClusterName    string
	ClusterLabels  map[string]string
	KubeDistro     string
	KubeconfigFile string
	KubectlCli     string
	MissionMatch   map[string]bool
}

//NewMissionManager creates new mission manager object
func NewMissionManager(edgeClusterConfig *v1alpha1.Clusterd) *MissionManager {

	// No need to check the config, as it was checked during the registration
	basedir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	return &MissionManager{
		ClusterName:    edgeClusterConfig.Name,
		ClusterLabels:  edgeClusterConfig.Labels,
		KubeDistro:     edgeClusterConfig.KubeDistro,
		KubeconfigFile: edgeClusterConfig.Kubeconfig,
		KubectlCli:     filepath.Join(basedir, DistroToKubectl[edgeClusterConfig.KubeDistro]),
		MissionMatch:   map[string]bool{},
	}
}

func (m *MissionManager) ApplyMission(mission *edgeclustersv1.Mission) error {
	m.MissionMatch[mission.Name] = m.isMatchingMission(mission)

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
			klog.Errorf("Mission %v is saved.", mission.Name)
		}
	}

	if m.isMatchingMission(mission) == false {
		klog.Infof("Mission %v does not match this cluster, skip the content applying", mission.Name)
		return nil
	}

	deploy_content_cmd := fmt.Sprintf("printf \"%s\" | %s apply --kubeconfig=%s -f - ", mission.Spec.Content, m.KubectlCli, m.KubeconfigFile)
	if _, err := ExecCommandLine(deploy_content_cmd, COMMAND_TIMEOUT_SEC); err != nil {
		return fmt.Errorf("Failed to apply the content of mission %v: %v", mission.Name, err)
	}

	klog.Infof("The content of mission %v applied successfully ", mission.Name)

	return nil
}

func (m *MissionManager) DeleteMission(mission *edgeclustersv1.Mission) error {
	delete(m.MissionMatch, mission.Name)
	if m.isMatchingMission(mission) == false {
		klog.Infof("Mission %v does not match this cluster", mission.Name)
	} else {
		delete_content_cmd := fmt.Sprintf("printf \"%s\" | %s delete --kubeconfig=%s -f - ", mission.Spec.Content, m.KubectlCli, m.KubeconfigFile)
		_, err := ExecCommandLine(delete_content_cmd, COMMAND_TIMEOUT_SEC)
		if err != nil {
			klog.Errorf("Failed to revert the content of mission %v: %v", mission.Name, err)
		} else {
			klog.Errorf("The content of mission %v is reverted.", mission.Name)
		}
	}

	delete_mission_cmd := fmt.Sprintf("%s delete mission %s --kubeconfig=%s", m.KubectlCli, mission.Name, m.KubeconfigFile)
	if _, err := ExecCommandLine(delete_mission_cmd, COMMAND_TIMEOUT_SEC); err != nil {
		return fmt.Errorf("Failed to delete mission %v: %v", mission.Name, err)
	}

	klog.Infof("Mission %v deleted successfully ", mission.Name)

	return nil
}

func (m *MissionManager) DeleteMissionByName(name string) error {
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

func (m *MissionManager) GetLocalMissionNames() ([]string, error) {
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

func (m *MissionManager) isMatchingMission(mission *edgeclustersv1.Mission) bool {
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

func (m *MissionManager) AlignMissionList(missionList []*edgeclustersv1.Mission) error {
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
