package types

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"

	edgeclustersv1 "github.com/kubeedge/kubeedge/cloud/pkg/apis/edgeclusters/v1"
)

// PodStatusRequest is Message.Content which come from edge
type PodStatusRequest struct {
	UID    types.UID
	Name   string
	Status v1.PodStatus
}

//ExtendResource is extended resource details that come from edge
type ExtendResource struct {
	Name     string            `json:"name,omitempty"`
	Type     string            `json:"type,omitempty"`
	Capacity resource.Quantity `json:"capacity,omitempty"`
}

// NodeStatusRequest is Message.Content which come from edge
type NodeStatusRequest struct {
	UID             types.UID
	Status          v1.NodeStatus
	ExtendResources map[v1.ResourceName][]ExtendResource
}

// EdgeClusterStateRequest is Message.Content which come from edge
type EdgeClusterStateRequest struct {
	UID   types.UID
	State edgeclustersv1.EdgeClusterState
}

// MissionStateRequest is Message.Content which come from edge
type MissionStateRequest struct {
	UID         types.UID
	ClusterName string
	State       map[string]string
}
