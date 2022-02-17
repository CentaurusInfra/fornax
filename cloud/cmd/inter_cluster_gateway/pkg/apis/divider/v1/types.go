/*
Copyright 2022 Authors of Fornax.
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
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DividerSpec defines the desired subnets
type DividerSpec struct {
	CreateTime     int    `json:"createTime"`
	Droplet        string `json:"droplet"`
	IP             string `json:"ip"`
	Mac            string `json:"mac"`
	ProvisionDelay string `json:"provisiondelay"`
	Status         string `json:"status"`
	Vni            string `json:"vni"`
	Vpc            string `json:"vpc"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Divider is the Schema for the Subnet API
type Divider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DividerSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// SubnetList is a list of Subnet resources.
type DividerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Divider `json:"items"`
}
