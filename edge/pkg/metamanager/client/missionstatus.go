package client

import (
	"fmt"

	"github.com/kubeedge/beehive/pkg/core/model"
	edgeapi "github.com/kubeedge/kubeedge/common/types"
	"github.com/kubeedge/kubeedge/edge/pkg/common/message"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
)

//MissionStatusGetter is interface to get mission status
type MissionStatusGetter interface {
	MissionStatus(namespace string) MissionStatusInterface
}

//MissionStatusInterface is mission status interface
type MissionStatusInterface interface {
	Create(*edgeapi.MissionStatusRequest) (*edgeapi.MissionStatusRequest, error)
	Update(rsName string, ns edgeapi.MissionStatusRequest) error
	Delete(name string) error
	Get(name string) (*edgeapi.MissionStatusRequest, error)
}

type missionStatus struct {
	namespace string
	send      SendInterface
}

func newMissionStatus(namespace string, s SendInterface) *missionStatus {
	return &missionStatus{
		send:      s,
		namespace: namespace,
	}
}

func (c *missionStatus) Create(ns *edgeapi.MissionStatusRequest) (*edgeapi.MissionStatusRequest, error) {
	return nil, nil
}

func (c *missionStatus) Update(rsName string, ms edgeapi.MissionStatusRequest) error {
	resource := fmt.Sprintf("%s/%s/%s", c.namespace, model.ResourceTypeMissionStatus, rsName)
	missionStatusMsg := message.BuildMsg(modules.MetaGroup, "", modules.ClusterdModuleName, resource, model.UpdateOperation, ms)
	_, err := c.send.SendSync(missionStatusMsg)
	if err != nil {
		return fmt.Errorf("update missionStatus failed, err: %v", err)
	}

	return nil
}

func (c *missionStatus) Delete(name string) error {
	return nil
}

func (c *missionStatus) Get(name string) (*edgeapi.MissionStatusRequest, error) {
	return nil, nil
}
