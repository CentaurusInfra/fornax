/*
Copyright 2020 The KubeEdge Authors.

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

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Mission specifies a workload to deploy in edge clusters
type Mission struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MissionSpec `json:"spec"`

	State map[string]string `json:"state,omitempty"`
}

// MissionSpec is a description of Mission
type MissionSpec struct {
	MissionResource string `json:"missionresource,omitempty"`

	MissionCommand MissionCommandSpec `json:"missioncommand,omitempty"`

	Placement GenericPlacementFields `json:"placement,omitempty"`

	StateCheck StateCheckFields `json:"statecheck"`
}

// MissionSpec is a description of Mission
type MissionCommandSpec struct {
	// a  test to run before running the mission command
	Trigger string `json:"trigger,omitempty"`

	// default false, which means the mission command will run when the trigger command fails.
	// if it is true, the mission command will run when the trigger command succeeds.
	RunWhenTriggerSucceed bool `json:"runwhentriggersucceed"`

	Command string `json:"command,omitempty"`

	ReverseCommand string `json:"reversecommand,omitempty"`
}

type GenericClusterReference struct {
	Name string `json:"name"`
}

type GenericPlacementFields struct {
	Clusters    []GenericClusterReference `json:"clusters,omitempty"`
	MatchLabels map[string]string         `json:"matchlabels,omitempty"`
}

type StateCheckFields struct {
	Command string `json:"command,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MissionList is a list of Mission objects.
type MissionList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of Mission objects in the list
	Items []Mission `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EdgeCluster specifies an edge cluster
type EdgeCluster struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines desired state of network
	// +optional
	Spec EdgeClusterSpec `json:"spec"`

	State EdgeClusterState `json:"state,omitempty"`
}

// EdgeCluster indicates the edge cluster config
type EdgeClusterSpec struct {
	// kubeconfig indicates the path to the edge cluster kubeconfig file
	Kubeconfig string `json:"kubeconfig,omitempty"`

	// Distribution of the cluster, supported value: arkots, to support in the furture: k3s
	KubeDistro string `json:"kubeDistro,omitempty"`

	// labels of the cluster
	Labels map[string]string `json:"labels,omitempty"`
}

// EdgeClusterState is a description of the state of edgecluster
type EdgeClusterState struct {
	HealthStatus string `json:"healthstatus,omitempty"`

	SubEdgeClusterStates map[string]string `json:"subedgeclusterstates,omitempty"`

	Nodes []string `json:"nodes,omitempty"`

	EdgeNodes []string `json:"edgenodes,omitempty"`

	ReceivedMissions []string `json:"receivedmissions,omitempty"`

	ActiveMissions []string `json:"activemissions,omitempty"`

	LastHeartBeat metav1.Time `json:"lastheartbeat,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MissionList is a list of Mission objects.
type EdgeClusterList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of Mission objects in the list
	Items []EdgeCluster `json:"items"`
}
