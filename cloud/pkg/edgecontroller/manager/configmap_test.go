/*
Copyright 2021 The KubeEdge Authors.

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

package manager

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/kubeedge/kubeedge/cloud/pkg/common/client"
	"github.com/kubeedge/kubeedge/cloud/pkg/common/informers"
	"github.com/kubeedge/kubeedge/cloud/pkg/edgecontroller/config"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
)

func TestConfigMapManager_Events(t *testing.T) {
	type fields struct {
		events chan watch.Event
	}

	ch := make(chan watch.Event, 1)
	tests := []struct {
		name   string
		fields fields
		want   chan watch.Event
	}{
		{
			"TestConfigMapManager_Events(): Case 1",
			fields{
				events: ch,
			},
			ch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmm := &ConfigMapManager{
				events: tt.fields.events,
			}
			if got := cmm.Events(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConfigMapManager.Events() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewConfigMapManager(t *testing.T) {
	type args struct {
		informer cache.SharedIndexInformer
	}

	config.Config.Buffer = &v1alpha1.EdgeControllerBuffer{
		ConfigMapEvent: 1024,
	}

	tmpfile, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmpfile.Name())
	if err := ioutil.WriteFile(tmpfile.Name(), []byte(mockKubeConfigContent), 0666); err != nil {
		t.Error(err)
	}
	client.InitKubeEdgeClient(&v1alpha1.KubeAPIConfig{
		KubeConfig:  tmpfile.Name(),
		QPS:         100,
		Burst:       200,
		ContentType: "application/vnd.kubernetes.protobuf",
	})

	tests := []struct {
		name string
		args args
	}{
		{
			"TestNewConfigMapManager(): Case 1",
			args{
				informers.GetInformersManager().GetK8sInformerFactory().Core().V1().ConfigMaps().Informer(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NewConfigMapManager(tt.args.informer)
		})
	}
}

func TestAddOrUpdateConfigMaps(t *testing.T) {
	ch := make(chan watch.Event, 1)
	tests := []struct {
		name  string
		input *corev1.ConfigMap
	}{
		{
			"AddOrUpdateConfigMap Case 1",
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: corev1.SchemeGroupVersion.String(),
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "cluster-gateway-config",
					Namespace:       "default",
					ResourceVersion: "1",
					UID:             "0387947d-d637-4f92-965d-11b8bdc9f548",
				},
				Data: map[string]string{
					"gateway_host_ip": "172.31.11.31",
					"gateway_name":    "edge_gateway",
				},
				BinaryData: map[string][]byte{},
			},
		},
		{
			"AddOrUpdateConfigMap Case 2",
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: corev1.SchemeGroupVersion.String(),
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "cluster-gateway-config",
					Namespace:       "default",
					ResourceVersion: "2",
				},
				Data: map[string]string{
					"gateway_host_ip":   "127.0.0.1",
					"gateway_name":      "edge_gateway",
					"gateway_neighbors": "edge1_gateway=172.31.11.31,edge2-gateway=172.31.9.254",
				},
				BinaryData: map[string][]byte{},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmm := &ConfigMapManager{
				events: ch,
			}
			cmm.AddOrUpdateConfigMap(test.input)
			if got := cmm.GetConfigMap(test.input.Name); !reflect.DeepEqual(got, test.input) {
				t.Errorf("ConfigMapManager.AddOrUpdateConfigMap() = %v, want %v", got, test.input)
			}
		})
	}
}

func TestDeleteConfigMaps(t *testing.T) {
	ch := make(chan watch.Event, 1)
	tests := []struct {
		name  string
		input *corev1.ConfigMap
	}{
		{
			"TestDeleteConfigMaps Case 1",
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: corev1.SchemeGroupVersion.String(),
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "cluster-gateway-config1",
					Namespace:       "default",
					ResourceVersion: "1",
					UID:             "0387947d-d637-4f92-965d-11b8bdc9f548",
				},
				Data: map[string]string{
					"gateway_host_ip": "172.31.11.31",
					"gateway_name":    "edge_gateway",
				},
				BinaryData: map[string][]byte{},
			},
		},
		{
			"TestDeleteConfigMaps Case 2",
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: corev1.SchemeGroupVersion.String(),
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "cluster-gateway-config2",
					Namespace:       "default",
					ResourceVersion: "2",
				},
				Data: map[string]string{
					"gateway_host_ip":   "127.0.0.1",
					"gateway_name":      "edge_gateway",
					"gateway_neighbors": "edge1_gateway=172.31.11.31,edge2-gateway=172.31.9.254",
				},
				BinaryData: map[string][]byte{},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmm := &ConfigMapManager{
				events: ch,
			}
			cmm.AddOrUpdateConfigMap(test.input)
			cmm.DeleteConfigMap(test.input.Name)
			if got := cmm.GetConfigMap(test.input.Name); got != nil {
				t.Errorf("ConfigMapManager.DeleteConfigMap() failed to delete the configmap %s", got.Name)
			}
		})
	}
}
