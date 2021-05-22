package config

import (
	"sync"

	configv1alpha1 "github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
)

var Config Configure
var once sync.Once

type Configure struct {
	MissionStatePruner *configv1alpha1.MissionStatePruner
}

func InitConfigure(msp *configv1alpha1.MissionStatePruner) {
	once.Do(func() {
		Config = Configure{
			MissionStatePruner: msp,
		}
	})
}
