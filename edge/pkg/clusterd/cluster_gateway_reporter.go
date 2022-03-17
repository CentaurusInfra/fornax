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
	configMap, err := GetClusterGatewayConfigMap()
	if err != nil {
		return err
	}

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

func GetClusterGatewayConfigMap() (*v1.ConfigMap, error) {
	var res *v1.ConfigMap

	getClusterGatewayCmd := fmt.Sprintf(" %s get configmap cluster-gateway-config --kubeconfig=%s", config.Config.KubectlCli, config.Config.Kubeconfig)
	output, err := helper.ExecCommandToCluster(getClusterGatewayCmd)
	if err != nil {
		klog.Errorf("Failed to get cluster gateway: %v", err)
		return res, err
	}

	if strings.TrimSpace(output) == "" {
		klog.V(4).Infof("There is no cluster gateway config maps.")
		return res, err
	}

	if err := json.Unmarshal([]byte(output), res); err != nil {
		klog.Errorf("Error in unmarshall cluster gateway config map json: (%s), error: %v", output, err)
		return res, err
	}
	return res, nil
}
