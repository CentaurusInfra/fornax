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

package helper

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"k8s.io/klog/v2"

	edgeclustersv1 "github.com/kubeedge/kubeedge/cloud/pkg/apis/edgeclusters/v1"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/config"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/util"
)

func GetLocalClusterScopeResourceNames(resType string, label string) []string {
	labelOption := ""
	if len(label) > 0 {
		labelOption = "-l " + label
	}
	getResourceCmd := fmt.Sprintf(" %s get %s -o json %s --kubeconfig=%s | jq -r '.items[] | [.metadata.name] | @tsv' ", config.Config.KubectlCli, resType, labelOption, config.Config.Kubeconfig)
	output, err := ExecCommandToCluster(getResourceCmd)
	if err != nil {
		klog.Errorf("Failed to get %v: %v", resType, err)
		return []string{err.Error()}
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

func GetMissionByName(name string) (*edgeclustersv1.Mission, error) {
	getMissionCmd := fmt.Sprintf("%s get mission %s --kubeconfig=%s -o json ", config.Config.KubectlCli, name, config.Config.Kubeconfig)
	output, err := ExecCommandToCluster(getMissionCmd)
	if err != nil {
		return nil, fmt.Errorf("Failed to get mission %v: %v", name, err)
	}

	var mission edgeclustersv1.Mission
	err = json.Unmarshal([]byte(output), &mission)
	if err != nil {
		return nil, err
	}

	return &mission, nil
}

func TestClusterReady() bool {
	testClusterCommand := fmt.Sprintf("%s cluster-info --kubeconfig=%s", config.Config.KubectlCli, config.Config.Kubeconfig)
	if _, err := util.ExecCommandLine(testClusterCommand); err != nil {
		klog.Errorf("The cluster is unhealthy: %v", err)
		return false
	}

	return true
}

// this is a hacky way for PoC only
func AnalyzeMissionContent(content string) (kind string, name string, namespace string) {
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

func EqualClusterReferences(a []edgeclustersv1.GenericClusterReference, b []edgeclustersv1.GenericClusterReference) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil && len(b) == 0 {
		return true
	}
	if b == nil && len(a) == 0 {
		return true
	}

	// for other cases, use the regular array compare
	return reflect.DeepEqual(a, b)
}

func EqualMaps(a map[string]string, b map[string]string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil && len(b) == 0 {
		return true
	}
	if b == nil && len(a) == 0 {
		return true
	}

	// for other cases, use the regular map compare
	return reflect.DeepEqual(a, b)
}

// As Json encoder may turn an empty array into nil, two same mission spec object may no long deep-equal.
// This function aims to detect two spec objects are truly equal.
func EqualMissionSpec(a edgeclustersv1.MissionSpec, b edgeclustersv1.MissionSpec) bool {
	if strings.TrimSpace(a.MissionResource) != strings.TrimSpace(b.MissionResource) {
		return false
	}

	if strings.TrimSpace(a.MissionCommand.Trigger) != strings.TrimSpace(b.MissionCommand.Trigger) {
		return false
	}

	if a.MissionCommand.RunWhenTriggerSucceed != b.MissionCommand.RunWhenTriggerSucceed {
		return false
	}

	if strings.TrimSpace(a.MissionCommand.Command) != strings.TrimSpace(b.MissionCommand.Command) {
		return false
	}

	if strings.TrimSpace(a.MissionCommand.ReverseCommand) != strings.TrimSpace(b.MissionCommand.ReverseCommand) {
		return false
	}

	if strings.TrimSpace(a.StateCheck.Command) != strings.TrimSpace(b.StateCheck.Command) {
		return false
	}

	if !EqualMaps(a.Placement.MatchLabels, b.Placement.MatchLabels) {
		return false
	}

	if !EqualClusterReferences(a.Placement.Clusters, b.Placement.Clusters) {
		return false
	}

	return true
}

func ExecCommandToCluster(commandline string) (string, error) {
	if !TestClusterReady() {
		return "", fmt.Errorf("cluster unhealthy")
	}

	return util.ExecCommandLine(commandline)
}
