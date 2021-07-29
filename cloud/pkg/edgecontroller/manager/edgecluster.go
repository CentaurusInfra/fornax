package manager

import (
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/kubeedge/kubeedge/cloud/pkg/edgecontroller/config"
)

// EdgeClusterManager manage all events of rule by SharedInformer
type EdgeClusterManager struct {
	events chan watch.Event
}

// Events return the channel save events from watch secret change
func (rem *EdgeClusterManager) Events() chan watch.Event {
	return rem.events
}

// NewEdgeClusterManager create EdgeClusterManager by SharedIndexInformer
func NewEdgeClusterManager(si cache.SharedIndexInformer) (*EdgeClusterManager, error) {
	events := make(chan watch.Event, config.Config.Buffer.EdgeClustersEvent)
	rh := NewCommonResourceEventHandler(events)
	si.AddEventHandler(rh)

	return &EdgeClusterManager{events: events}, nil
}
