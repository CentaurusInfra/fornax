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

	corev1 "k8s.io/api/core/v1"
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
	gateway_name, gateway_host_ip, err := GetClusterGatewayNameAndHostIP()
	if err != nil {
		return err
	}
	configMap := &corev1.ConfigMap{}
	configMap.Data = make(map[string]string)
	configMap.Data["gateway_name"] = gateway_name
	configMap.Data["gateway_host_ip"] = gateway_host_ip
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
	klog.Infof("Starting edgecluster state reporter.")
	defer klog.Infof("Shutting down edgecluster state reporter")

	go utilwait.Until(reporter.syncClusterGatewayConfigMap, reporter.clusterGatewayUpdateInterval, utilwait.NeverStop)
}

func GetClusterGatewayNameAndHostIP() (string, string, error) {
	var gateway_name string
	var gateway_ip string

	getClusterGatewayDataPCmd := fmt.Sprintf(" %s get configmap cluster-gateway-config -o=jsonpath='{.data}' --kubeconfig=%s", config.Config.KubectlCli, config.Config.Kubeconfig)
	output, err := helper.ExecCommandToCluster(getClusterGatewayDataPCmd)
	if err != nil {
		klog.Errorf("Failed to get cluster gateway host ip: %v", err)
		return gateway_name, gateway_ip, err
	}

	if strings.TrimSpace(output) == "" {
		klog.V(4).Infof("There is no cluster gateway host ip.")
		return gateway_name, gateway_ip, err
	}

	var dataMap map[string]string

	if err := json.Unmarshal([]byte(output), &dataMap); err != nil {
		klog.Errorf("Error in unmarshall cluster data json: (%s), error: %v", output, err)
		return gateway_name, gateway_ip, err
	}
	return dataMap["gateway_name"], dataMap["gateway_host_ip"], nil
}

func (reporter *ClusterGatewayReporter) UnmarshalAndUpdateNeighbors(content []byte) (err error) {
	var lists []string
	if err = json.Unmarshal(content, &lists); err != nil {
		return err
	}
	if existingConfigMap, err := reporter.clusterd.metaClient.ConfigMaps(reporter.clusterd.namespace).Get(constants.ClusterGatewayConfigMap); err != nil {
		for _, list := range lists {
			var configMap v1.ConfigMap
			err = json.Unmarshal([]byte(list), &configMap)
			if err != nil {
				return err
			}
			if _, updated, err := util.GetUpdatedClusterGatewayNeighbors(configMap.Data["gateway_name"], configMap.Data["gateway_host_ip"], existingConfigMap); updated && err == nil {
				if err := reporter.clusterd.metaClient.ConfigMaps(reporter.clusterd.namespace).Update(existingConfigMap); err != nil {
					return err
				}
			} else {
				return err
			}

		}
	} else {
		return err
	}

	return nil

}
