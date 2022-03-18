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
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/config"
	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/helper"
)

type ClusterGatewayReporter struct {
	clusterd                     *clusterd
	clusterGatewayUpdateInterval time.Duration
}

func NewClusterGatewayReporter(c *clusterd) *ClusterGatewayReporter {
	return &ClusterGatewayReporter{
		clusterd:                     c,
		clusterGatewayUpdateInterval: 60 * time.Second,
	}
}

func (reporter *ClusterGatewayReporter) updateClusterGatewayConfigMap() error {
	clusterGatewayHostIP, err := GetClusterGatewayHostIP()
	if err != nil {
		return err
	}
	configMap := &corev1.ConfigMap{}
	configMap.Data = make(map[string]string)
	configMap.Data["gateway_host_ip"] = clusterGatewayHostIP
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
		klog.Errorf("Unable to update cluster gateway config maps: %v", err)
	}
}

func (reporter *ClusterGatewayReporter) Run() {
	klog.Infof("Starting edgecluster state reporter.")
	defer klog.Infof("Shutting down edgecluster state reporter")

	go utilwait.Until(reporter.syncClusterGatewayConfigMap, reporter.clusterGatewayUpdateInterval, utilwait.NeverStop)
}

func GetClusterGatewayHostIP() (string, error) {
	var res string

	getClusterGatewayHostIPCmd := fmt.Sprintf(" %s get configmap cluster-gateway-config -o=jsonpath='{.data.gateway_host_ip}' --kubeconfig=%s", config.Config.KubectlCli, config.Config.Kubeconfig)
	output, err := helper.ExecCommandToCluster(getClusterGatewayHostIPCmd)
	if err != nil {
		klog.Errorf("Failed to get cluster gateway host ip: %v", err)
		return res, err
	}

	if strings.TrimSpace(output) == "" {
		klog.V(4).Infof("There is no cluster gateway host ip.")
		return res, err
	}

	res = string(output)
	return res, nil
}
