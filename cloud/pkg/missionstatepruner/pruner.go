package missionstatepruner

import (
	"context"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/kubeedge/beehive/pkg/core"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	crdClientset "github.com/kubeedge/kubeedge/cloud/pkg/client/clientset/versioned"
	keclient "github.com/kubeedge/kubeedge/cloud/pkg/common/client"
	"github.com/kubeedge/kubeedge/cloud/pkg/common/modules"
	"github.com/kubeedge/kubeedge/cloud/pkg/missionstatepruner/config"
	configv1alpha1 "github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
)

const (
	EdgeClusterOffline = "cluster unreacheable"
	HealthyStatus      = "healthy"
	DisconnectedStatus = "disconnected"
)

//Mission state pruner periodically update the mission state when some edgeclusters become offline
type MissionStatePruner struct {
	enable             bool
	syncInterval       time.Duration
	edgeclusterTimeout time.Duration
	// Here we use a client, instead of an informer/lister, to check whether edgeclusters are offline.
	// Reason I: We check the edge cluster once for a long period ( default one per minute), to avoid noise of edge cluster state flip&flop.
	// So a long connection used by informer/lister actually is more expensive.
	// Reason II: informers are good at detecting object changes. But here we are interested in edgeclusters whose status have NOT changed for a long time.
	crdClient             crdClientset.Interface
	deadEdgeClustersCache map[string]bool
}

func newMissionStatePruner(msp *configv1alpha1.MissionStatePruner) *MissionStatePruner {
	return &MissionStatePruner{
		enable:                msp.Enable,
		crdClient:             keclient.GetCRDClient(),
		syncInterval:          time.Duration(msp.SyncInterval) * time.Second,
		edgeclusterTimeout:    time.Duration(msp.EdgeClusterTimeout) * time.Second,
		deadEdgeClustersCache: map[string]bool{},
	}
}

func Register(msp *configv1alpha1.MissionStatePruner) {
	config.InitConfigure(msp)
	core.Register(newMissionStatePruner(msp))
}

func (msp *MissionStatePruner) Name() string {
	return modules.MissionStatePrunerModuleName
}

func (msp *MissionStatePruner) Group() string {
	return modules.SyncControllerModuleGroup
}

func (msp *MissionStatePruner) Enable() bool {
	return msp.enable
}

func (msp *MissionStatePruner) Start() {
	go wait.Until(msp.checkAndPrune, msp.syncInterval, beehiveContext.Done())
}

func (msp *MissionStatePruner) checkAndPrune() {
	allEdgeClusters, err := msp.crdClient.EdgeclustersV1().EdgeClusters().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Error in listing edge clusters: %v", err)
		return
	}

	deadEdgeClusters := map[string]bool{}
	newDeadEdgeClusters := false
	for _, ec := range allEdgeClusters.Items {
		if time.Since(ec.State.LastHeartBeat.Time) > msp.edgeclusterTimeout {
			if ec.State.HealthStatus != DisconnectedStatus {
				ec.State.HealthStatus = DisconnectedStatus
				ec.State.SubEdgeClusterStates = map[string]string{}
				ec.State.Nodes = []string{}
				ec.State.EdgeNodes = []string{}
				ec.State.ReceivedMissions = []string{}
				ec.State.ActiveMissions = []string{}

				_, err := msp.crdClient.EdgeclustersV1().EdgeClusters().Update(context.Background(), &ec, metav1.UpdateOptions{})
				if err != nil {
					// if there is an error, log it and move on.
					klog.Warningf("Failed to update the status of EdgeCluster %s to Disconnected: %v", ec.Name, err)
				}
			}
		}

		if ec.State.HealthStatus != HealthyStatus {
			deadEdgeClusters[ec.Name] = true
			if _, ok := msp.deadEdgeClustersCache[ec.Name]; !ok {
				newDeadEdgeClusters = true
			}
		}
	}

	msp.deadEdgeClustersCache = deadEdgeClusters
	if !newDeadEdgeClusters {
		klog.V(4).Infof("Did not find new offline edge clusters.")
		return
	}

	allMissions, err := msp.crdClient.EdgeclustersV1().Missions().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Error in listing missions: %v", err)
		return
	}

	for _, mission := range allMissions.Items {
		changed := false

		for key, val := range mission.State {
			parts := strings.Split(key, "/")
			switch len(parts) {
			case 0:
				// it should not happen, an empty mission names should be blocked at creation
				klog.Errorf("invalid mission state key.")
				continue
			case 1:
				// set the mission state about the dead edgecluster
				if deadEdgeClusters[key] && val != EdgeClusterOffline {
					mission.State[key] = EdgeClusterOffline
					changed = true
				}
			default:
				// remove the mission state about the sub-clusters of the dead edgecluster
				if deadEdgeClusters[parts[0]] {
					delete(mission.State, key)
					changed = true
				}
			}
		}
		if changed {
			klog.V(3).Infof("Updating the state of mission %v ...", mission.Name)
			_, err := msp.crdClient.EdgeclustersV1().Missions().Update(context.Background(), &mission, metav1.UpdateOptions{})
			if err != nil {
				klog.Warningf("Error in updating the state of mission %v: %v", mission.Name, err)
			}
		}
	}
}
