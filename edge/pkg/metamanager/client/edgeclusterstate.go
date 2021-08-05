package client

import (
	"fmt"

	"github.com/kubeedge/beehive/pkg/core/model"
	edgeapi "github.com/kubeedge/kubeedge/common/types"
	"github.com/kubeedge/kubeedge/edge/pkg/common/message"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
)

//EdgeClusterStateGetter is interface to get edgeCluster status
type EdgeClusterStateGetter interface {
	EdgeClusterState(namespace string) EdgeClusterStateInterface
}

//EdgeClusterStateInterface is edgeCluster status interface
type EdgeClusterStateInterface interface {
	Create(*edgeapi.EdgeClusterStateRequest) (*edgeapi.EdgeClusterStateRequest, error)
	Update(rsName string, ns edgeapi.EdgeClusterStateRequest) error
	Delete(name string) error
	Get(name string) (*edgeapi.EdgeClusterStateRequest, error)
}

type edgeClusterState struct {
	namespace string
	send      SendInterface
}

func newEdgeClusterState(namespace string, s SendInterface) *edgeClusterState {
	return &edgeClusterState{
		send:      s,
		namespace: namespace,
	}
}

func (c *edgeClusterState) Create(ns *edgeapi.EdgeClusterStateRequest) (*edgeapi.EdgeClusterStateRequest, error) {
	return nil, nil
}

func (c *edgeClusterState) Update(rsName string, ecs edgeapi.EdgeClusterStateRequest) error {
	resource := fmt.Sprintf("%s/%s/%s", c.namespace, model.ResourceTypeEdgeClusterState, rsName)
	edgeClusterStateMsg := message.BuildMsg(modules.MetaGroup, "", modules.ClusterdModuleName, resource, model.UpdateOperation, ecs)
	_, err := c.send.SendSync(edgeClusterStateMsg)
	if err != nil {
		return fmt.Errorf("update edgeClusterState failed, err: %v", err)
	}

	return nil
}

func (c *edgeClusterState) Delete(name string) error {
	return nil
}

func (c *edgeClusterState) Get(name string) (*edgeapi.EdgeClusterStateRequest, error) {
	return nil, nil
}
