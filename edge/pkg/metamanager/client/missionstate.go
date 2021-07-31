package client

import (
	"fmt"

	"github.com/kubeedge/beehive/pkg/core/model"
	edgeapi "github.com/kubeedge/kubeedge/common/types"
	"github.com/kubeedge/kubeedge/edge/pkg/common/message"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
)

//MissionStateGetter is interface to get mission status
type MissionStateGetter interface {
	MissionState(namespace string) MissionStateInterface
}

//MissionStateInterface is mission status interface
type MissionStateInterface interface {
	Create(*edgeapi.MissionStateRequest) (*edgeapi.MissionStateRequest, error)
	Update(rsName string, ns edgeapi.MissionStateRequest) error
	Delete(name string) error
	Get(name string) (*edgeapi.MissionStateRequest, error)
}

type missionState struct {
	namespace string
	send      SendInterface
}

func newMissionState(namespace string, s SendInterface) *missionState {
	return &missionState{
		send:      s,
		namespace: namespace,
	}
}

func (c *missionState) Create(ns *edgeapi.MissionStateRequest) (*edgeapi.MissionStateRequest, error) {
	return nil, nil
}

func (c *missionState) Update(rsName string, ms edgeapi.MissionStateRequest) error {
	resource := fmt.Sprintf("%s/%s/%s", c.namespace, model.ResourceTypeMissionState, rsName)
	missionStateMsg := message.BuildMsg(modules.MetaGroup, "", modules.ClusterdModuleName, resource, model.UpdateOperation, ms)
	_, err := c.send.SendSync(missionStateMsg)
	if err != nil {
		return fmt.Errorf("update missionState failed, err: %v", err)
	}

	return nil
}

func (c *missionState) Delete(name string) error {
	return nil
}

func (c *missionState) Get(name string) (*edgeapi.MissionStateRequest, error) {
	return nil, nil
}
