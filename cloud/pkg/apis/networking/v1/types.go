/*
Copyright 2020 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing pervpcs and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Vpc specifies a VPC object
type Vpc struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VpcSpec `json:"spec"`

	State map[string]string `json:"state,omitempty"`
}

// VpcSpec is a description of Vpc
type VpcSpec struct {
	CIDR string `json:"cidr,omitempty"`

	Prefix string `json:"prefix,omitempty"`

	Vni string `json:"vni,omitempty"`

	Subnets []Subnet `json:"subnets,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VpcList is a list of Vpc objects.
type VpcList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of Vpc objects in the list
	Items []Vpc `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Subnet specifies a subnet
type Subnet struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines desired state of network
	// +optional
	Spec SubnetSpec `json:"spec"`

	Status SubnetStatus `json:"status,omitempty"`
}

// SubnetSpec indicates the Subnet config
type SubnetSpec struct {
	CIDR string `json:"cidr,omitempty"`

	Prefix string `json:"prefix,omitempty"`

	Vni string `json:"vni,omitempty"`

	Vpc string `json:"vpc,omitempty"`
}

// SubnetStatus is a description of Vpc status
type SubnetStatus struct {
	Healthy bool `json:"healthy,omitempty"`

	FreeIps int `json:"freeips,omitempty"`

	UsedIps int `json:"usedips,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VpcList is a list of Vpc objects.
type SubnetList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of Vpc objects in the list
	Items []Subnet `json:"items"`
}
