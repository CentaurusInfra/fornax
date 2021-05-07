package client

import (
	"fmt"

	"github.com/kubeedge/beehive/pkg/core/model"
	edgeapi "github.com/kubeedge/kubeedge/common/types"
	"github.com/kubeedge/kubeedge/edge/pkg/common/message"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
)

//EdgeClusterStatusGetter is interface to get edgeCluster status
type EdgeClusterStatusGetter interface {
	EdgeClusterStatus(namespace string) EdgeClusterStatusInterface
}

//EdgeClusterStatusInterface is edgeCluster status interface
type EdgeClusterStatusInterface interface {
	Create(*edgeapi.EdgeClusterStatusRequest) (*edgeapi.EdgeClusterStatusRequest, error)
	Update(rsName string, ns edgeapi.EdgeClusterStatusRequest) error
	Delete(name string) error
	Get(name string) (*edgeapi.EdgeClusterStatusRequest, error)
}

type edgeClusterStatus struct {
	namespace string
	send      SendInterface
}

func newEdgeClusterStatus(namespace string, s SendInterface) *edgeClusterStatus {
	return &edgeClusterStatus{
		send:      s,
		namespace: namespace,
	}
}

func (c *edgeClusterStatus) Create(ns *edgeapi.EdgeClusterStatusRequest) (*edgeapi.EdgeClusterStatusRequest, error) {
	return nil, nil
}

func (c *edgeClusterStatus) Update(rsName string, ecs edgeapi.EdgeClusterStatusRequest) error {
	resource := fmt.Sprintf("%s/%s/%s", c.namespace, model.ResourceTypeEdgeClusterStatus, rsName)
	edgeClusterStatusMsg := message.BuildMsg(modules.MetaGroup, "", modules.EdgeClusterModuleName, resource, model.UpdateOperation, ecs)
	_, err := c.send.SendSync(edgeClusterStatusMsg)
	if err != nil {
		return fmt.Errorf("update edgeClusterStatus failed, err: %v", err)
	}

	return nil
}

func (c *edgeClusterStatus) Delete(name string) error {
	return nil
}

func (c *edgeClusterStatus) Get(name string) (*edgeapi.EdgeClusterStatusRequest, error) {
	return nil, nil
}
