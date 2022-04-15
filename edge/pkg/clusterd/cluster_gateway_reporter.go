/*
Copyright 2022 The Kubernetes Authors.

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
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/kubeedge/kubeedge/common/constants"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/config"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/helper"
	"github.com/kubeedge/kubeedge/pkg/util"
)

type ClusterGatewayReporter struct {
	clusterd                     *clusterd
	clusterGatewayUpdateInterval time.Duration
}

func NewClusterGatewayReporter(c *clusterd) *ClusterGatewayReporter {
	return &ClusterGatewayReporter{
		clusterd:                     c,
		clusterGatewayUpdateInterval: time.Duration(config.Config.ClusterGatewayUpdateInterval) * time.Second,
	}
}

func (reporter *ClusterGatewayReporter) updateClusterGatewayConfigMap() error {
	gatewayName, gatewayHostIP, err := GetClusterGatewayNameAndHostIP()
	if err != nil {
		return err
	}
	configMap := &v1.ConfigMap{}
	configMap.Data = make(map[string]string)
	configMap.Data[constants.ClusterGatewayConfigMapGatewayName] = gatewayName
	configMap.Data[constants.ClusterGatewayConfigMapGatewayHostIP] = gatewayHostIP
	configMap.ClusterName = config.Config.Name
	err = reporter.clusterd.metaClient.ConfigMaps(reporter.clusterd.namespace).Update(configMap)
	if err != nil {
		klog.Errorf("update cluster gateway config map failed, error: %v", err)
		return err
	}
	return nil
}

func (reporter *ClusterGatewayReporter) syncClusterGatewayConfigMap() {
	if err := reporter.updateClusterGatewayConfigMap(); err != nil {
		klog.Errorf("unable to update cluster gateway config maps: %v", err)
	}
}

func (reporter *ClusterGatewayReporter) Run() {
	klog.Infof("starting cluster gateway reporter.")
	defer klog.Infof("shutting down cluster gateway reporter")

	go utilwait.Until(reporter.syncClusterGatewayConfigMap, reporter.clusterGatewayUpdateInterval, utilwait.NeverStop)
}

func GetClusterGatewayNameAndHostIP() (string, string, error) {
	var gatewayName string
	var gatewayIP string

	getClusterGatewayDataPCmd := fmt.Sprintf(" %s get configmap cluster-gateway-config -o=jsonpath='{.data}' --kubeconfig=%s", config.Config.KubectlCli, config.Config.Kubeconfig)
	output, err := helper.ExecCommandToCluster(getClusterGatewayDataPCmd)
	if err != nil {
		klog.Errorf("failed to get cluster gateway host ip: %v", err)
		return gatewayName, gatewayIP, err
	}

	if strings.TrimSpace(output) == "" {
		return gatewayName, gatewayIP, err
	}

	var dataMap map[string]string

	if err := json.Unmarshal([]byte(output), &dataMap); err != nil {
		klog.Errorf("error in unmarshall cluster data json: (%s), error: %v", output, err)
		return gatewayName, gatewayIP, err
	}
	return dataMap[constants.ClusterGatewayConfigMapGatewayName], dataMap[constants.ClusterGatewayConfigMapGatewayHostIP], nil
}

func (reporter *ClusterGatewayReporter) GetClusterGatewayNeighbors() (string, error) {
	var neighbors string

	getNeighborsCmd := fmt.Sprintf(" %s get configmap cluster-gateway-config -o=jsonpath='{.data.gateway_neighbors}' --kubeconfig=%s", config.Config.KubectlCli, config.Config.Kubeconfig)
	output, err := helper.ExecCommandToCluster(getNeighborsCmd)
	if err != nil {
		klog.Errorf("failed to get cluster gateway host ip: %v", err)
		return neighbors, err
	}
	if strings.TrimSpace(output) == "" {
		return "", nil
	}
	return string(output), nil
}

func (reporter *ClusterGatewayReporter) UnmarshalAndUpdateNeighbors(content []byte) (err error) {
	var lists []string
	if err = json.Unmarshal(content, &lists); err != nil {
		return err
	}
	for _, list := range lists {
		var configMap v1.ConfigMap
		err = json.Unmarshal([]byte(list), &configMap)
		if err != nil {
			return err
		}
		if err = reporter.UpdateNeighbor(configMap.Data[constants.ClusterGatewayConfigMapGatewayName], configMap.Data[constants.ClusterGatewayConfigMapGatewayHostIP]); err != nil {
			return err
		}
	}
	return nil
}

func (reporter *ClusterGatewayReporter) UpdateNeighbor(gatewayName, gatewayHost string) (err error) {
	neighbors, err := reporter.GetClusterGatewayNeighbors()
	if err != nil {
		return err
	}
	if updatedNeighbors, updated, err := util.GetUpdatedClusterGatewayNeighbors(gatewayName, gatewayHost, neighbors); updated && err == nil {
		neighborUpdateCommand := fmt.Sprintf("%s patch configmap %s --kubeconfig=%s --patch '{\"data\":{\"gateway_neighbors\":\"%s\"}}' --type=merge", config.Config.KubectlCli, constants.ClusterGatewayConfigMap, config.Config.Kubeconfig, updatedNeighbors)
		if _, err = helper.ExecCommandToCluster(neighborUpdateCommand); err != nil {
			if strings.Contains(err.Error(), "Error from server (NotFound):") {
				klog.Infof("configMap %v is deleted.", constants.ClusterGatewayConfigMap)
				return nil
			}
			klog.Errorf("error when checking the configmap %v with the error: %v", constants.ClusterGatewayConfigMap, err)
		}
	} else {
		klog.Errorf("error when get the configmap %v with the error: %v", constants.ClusterGatewayConfigMap, err)
	}
	return nil
}
