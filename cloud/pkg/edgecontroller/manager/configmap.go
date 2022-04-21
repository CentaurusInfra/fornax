package manager

import (
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/kubeedge/kubeedge/cloud/pkg/edgecontroller/config"
)

// ConfigMapManager manage all events of configmap by SharedInformer
type ConfigMapManager struct {
	events     chan watch.Event
	configMaps sync.Map
}

// Events return the channel save events from watch configmap change
func (cmm *ConfigMapManager) Events() chan watch.Event {
	return cmm.events
}

// NewConfigMapManager create ConfigMapManager by kube clientset and namespace
func NewConfigMapManager(si cache.SharedIndexInformer) (*ConfigMapManager, error) {
	events := make(chan watch.Event, config.Config.Buffer.ConfigMapEvent)
	rh := NewCommonResourceEventHandler(events)
	si.AddEventHandler(rh)

	return &ConfigMapManager{events: events}, nil
}

// AddOrUpdateConfigMap is to maintain the configMapin the cache up-to-date
func (cmm *ConfigMapManager) AddOrUpdateConfigMap(cm *v1.ConfigMap) {
	cmm.configMaps.Store(cm.Name, cm)
}

// Delete ConfigMap from cache
func (cmm *ConfigMapManager) DeleteConfigMap(cmName string) {
	cmm.configMaps.Delete(cmName)
}

// Get ConfigMap from cache
func (cmm *ConfigMapManager) GetConfigMap(cmName string) *v1.ConfigMap {
	if value, ok := cmm.configMaps.Load(cmName); ok {
		return value.(*v1.ConfigMap)
	}
	return nil
}
