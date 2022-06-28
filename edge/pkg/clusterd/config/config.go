package config

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"github.com/kubeedge/kubeedge/edge/pkg/clusterd/util"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/edgecore/v1alpha1"
)

var Config Configure
var once sync.Once

var DistroToKubectl = map[string]string{}

func init() {
	initializeDistroToKubectlMap()
}

func initializeDistroToKubectlMap() {
	basedir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	DistroToKubectl["arktos"] = filepath.Join(basedir, "kubectl/arktos/kubectl")
	DistroToKubectl["k8s"] = filepath.Join(basedir, "kubectl/vanilla/kubectl")
	DistroToKubectl["k3s"] = "sudo kubectl"
}

type Configure struct {
	v1alpha1.Clusterd
	KubectlCli string
}

func InitConfigure(c *v1alpha1.Clusterd) {
	c.Kubeconfig = v1alpha1.NewEdgeClusterEdgeCoreConfig().Modules.Clusterd.Kubeconfig
	once.Do(func() {
		Config = Configure{
			Clusterd: *c,
		}
		log.Println("verifying c.kubeconfig : ", c.Kubeconfig)
		if !util.FileExists(c.Kubeconfig) {
			klog.Fatalf("Could not open kubeconfig file (%s)", c.Kubeconfig)
		}

		if _, exists := DistroToKubectl[c.KubeDistro]; !exists {
			klog.Fatalf("Invalid kube distribution (%v)", c.KubeDistro)
		}

		Config.KubectlCli = DistroToKubectl[c.KubeDistro]

		// Kubernetes built-in labels
		Config.Labels[v1.LabelHostname] = c.Name
		// KubeEdge specific labels
		Config.Labels["role.kubernetes.io/edgecluster"] = ""
		Config.Labels["edgeclusters.kubeedge.io/kubedistro"] = c.KubeDistro
	})
}
