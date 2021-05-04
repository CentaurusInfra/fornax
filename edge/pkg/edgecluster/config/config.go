package config

import (
	"sync"

	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/edgecore/v1alpha1"
)

var Config Configure
var once sync.Once

type Configure struct {
	v1alpha1.EdgeCluster
}

func InitConfigure(e *v1alpha1.EdgeCluster) {
	once.Do(func() {
		Config = Configure{
			EdgeCluster: *e,
		}
	})
}
