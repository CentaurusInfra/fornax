package util

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	dividerclientset "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/divider/client/clientset/versioned"
	dividerv1 "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/divider/v1"
	subnetclientset "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/subnet/client/clientset/versioned"
	subnetinterface "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/subnet/client/clientset/versioned/typed/subnet/v1"
	subnetv1 "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/subnet/v1"
	vpcclientset "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/vpc/client/clientset/versioned"
	vpcv1 "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/vpc/v1"
	"github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/srv"
	"github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/srv/proto"
)

type KubeWatcher struct {
	kubeconfig               *rest.Config
	quit                     chan struct{}
	subnetInformer           cache.SharedIndexInformer
	vpcInformer              cache.SharedIndexInformer
	dividerInformer          cache.SharedIndexInformer
	gatewayconfigmapInformer cache.SharedIndexInformer
	subnetMap                map[string]*subnetv1.Subnet
	dividerMap               map[string]*dividerv1.Divider
	vpcMap                   map[string]*vpcv1.Vpc
	vpcGatewayMap            map[string]GatewayConfig
	gatewayMap               map[string]string
	subnetInterface          subnetinterface.SubnetInterface
	gatewayName              string
	gatewayHostIP            string
	client                   *srv.Client
}

func NewKubeWatcher(kubeconfig *rest.Config, client *srv.Client, quit chan struct{}) (*KubeWatcher, error) {
	gatewayconfigmapclientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	gatewayconfigmapSelector := fields.ParseSelectorOrDie("metadata.name=cluster-gateway-config")
	gatewayconfigmapLW := cache.NewListWatchFromClient(gatewayconfigmapclientset.CoreV1().RESTClient(), "configmaps", v1.NamespaceAll, gatewayconfigmapSelector)
	gatewayconfigmapInformer := cache.NewSharedIndexInformer(gatewayconfigmapLW, &v1.ConfigMap{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	subnetclientset, err := subnetclientset.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	subnetSelector := fields.ParseSelectorOrDie("metadata.name!=net0 ")
	subnetLW := cache.NewListWatchFromClient(subnetclientset.MizarV1().RESTClient(), "subnets", v1.NamespaceAll, subnetSelector)
	subnetInformer := cache.NewSharedIndexInformer(subnetLW, &subnetv1.Subnet{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	vpcclientset, err := vpcclientset.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	vpcSelector := fields.ParseSelectorOrDie("metadata.name!=vpc0")
	vpcLW := cache.NewListWatchFromClient(vpcclientset.MizarV1().RESTClient(), "vpcs", v1.NamespaceAll, vpcSelector)
	vpcInformer := cache.NewSharedIndexInformer(vpcLW, &vpcv1.Vpc{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	dividerclientset, err := dividerclientset.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	dividerLW := cache.NewListWatchFromClient(dividerclientset.MizarV1().RESTClient(), "dividers", v1.NamespaceAll, fields.Everything())
	dividerInformer := cache.NewSharedIndexInformer(dividerLW, &dividerv1.Divider{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	return &KubeWatcher{
		kubeconfig:               kubeconfig,
		quit:                     quit,
		gatewayconfigmapInformer: gatewayconfigmapInformer,
		subnetInformer:           subnetInformer,
		vpcInformer:              vpcInformer,
		dividerInformer:          dividerInformer,
		subnetMap:                make(map[string]*subnetv1.Subnet),
		dividerMap:               make(map[string]*dividerv1.Divider),
		vpcMap:                   make(map[string]*vpcv1.Vpc),
		vpcGatewayMap:            make(map[string]GatewayConfig),
		gatewayMap:               make(map[string]string),
		subnetInterface:          subnetclientset.MizarV1().Subnets("default"),
		client:                   client,
	}, nil
}

func (watcher *KubeWatcher) Run() {
	watcher.gatewayconfigmapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if cm, ok := obj.(*v1.ConfigMap); ok {
				watcher.gatewayName = cm.Data["gateway_name"]
				watcher.gatewayHostIP = cm.Data["gateway_host_ip"]
				gatewayArr := strings.Split(cm.Data["gateway_neighbors"], ",")
				for _, gatewayPair := range gatewayArr {
					gateway := strings.Split(gatewayPair, "=")
					watcher.gatewayMap[gateway[0]] = gateway[1]
				}
				klog.V(3).Infof("The gateway map is set to %v", watcher.gatewayMap)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			oldCm, newCm := old.(*v1.ConfigMap), new.(*v1.ConfigMap)
			if oldCm.Data["gateway_neighbors"] != newCm.Data["gateway_neighbors"] {
				klog.V(3).Infof("The new gateway neighbors are set to %v", newCm.Data["gateway_neighbors"])
				watcher.gatewayMap = make(map[string]string)
				gatewayArr := strings.Split(newCm.Data["gateway_neighbors"], ",")
				for _, gatewayPair := range gatewayArr {
					gateway := strings.Split(gatewayPair, "=")
					watcher.gatewayMap[gateway[0]] = gateway[1]
				}
				klog.V(3).Infof("The gateway map is updated to %v", watcher.gatewayMap)
			}
		},
	})
	watcher.subnetInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if subnet, ok := obj.(*subnetv1.Subnet); ok {
				klog.V(3).Infof("A new subnet %s is created", subnet.Name)
				if subnet.Spec.Virtual {
					return
				}
				for gatewayName, gatewayHostIP := range watcher.gatewayMap {
					klog.V(3).Infof("A new vpc %s is trying to sync to %s with the ip %s", subnet.Name, gatewayName, gatewayHostIP)
					conn, client, ctx, cancel, err := watcher.client.Connect(gatewayHostIP)
					if err != nil {
						klog.Errorf("failed to sync vpc to %s with the error %v", gatewayHostIP, err)
					}
					defer conn.Close()
					defer cancel()
					request := &proto.CreateSubnetRequest{
						Name: subnet.Name, Namespace: subnet.Namespace, IP: subnet.Spec.IP, Status: subnet.Spec.Status,
						Prefix: subnet.Spec.Prefix, Vpc: subnet.Spec.Vpc, Vni: subnet.Spec.Vni}
					returnMessage, err := client.CreateSubnet(ctx, request)
					klog.V(3).Infof("The returnMessage is %v", returnMessage)
					if err != nil {
						klog.Errorf("return from %s with the message %v with the error %v", gatewayHostIP, returnMessage, err)
					}
				}
			}
		},
	})

	watcher.vpcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if vpc, ok := obj.(*vpcv1.Vpc); ok {
				klog.V(3).Infof("A new vpc %s is created", vpc.Name)
				for gatewayName, gatewayHostIP := range watcher.gatewayMap {
					klog.V(3).Infof("A new vpc %s is trying to sync to %s with the ip %s", vpc.Name, gatewayName, gatewayHostIP)
					conn, client, ctx, cancel, err := watcher.client.Connect(gatewayHostIP)
					if err != nil {
						klog.Errorf("failed to sync vpc to %s with the error %v", gatewayHostIP, err)
					}
					defer conn.Close()
					defer cancel()
					request := &proto.CreateVpcGatewayRequest{Name: vpc.Name, Namespace: vpc.Namespace, GatewayName: watcher.gatewayName, GatewayHostIP: watcher.gatewayHostIP}
					returnMessage, err := client.CreateVpcGateway(ctx, request)
					klog.V(3).Infof("The returnMessage is %v", returnMessage)
					if err != nil {
						klog.Errorf("return from %s with the message %v with the error %v", gatewayHostIP, returnMessage, err)
					}
				}
			}
		},
	})

	watcher.dividerInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if divider, ok := obj.(*dividerv1.Divider); ok {
				klog.V(3).Infof("A new divider %s is added", divider.Name)
				if _, existed := watcher.dividerMap[divider.Name]; !existed {
					watcher.dividerMap[divider.Name] = divider
					if gateway, existed := watcher.vpcGatewayMap[divider.Spec.Vni]; existed {
						localDividerIP := net.ParseIP(divider.Spec.IP)
						if localDividerIP == nil {
							klog.Fatalf("Invalid local divider IP: %v", divider.Spec.IP)
						}
						localDividerIP = localDividerIP.To4()
						localDividerHosts := gateway.LocalDividerHosts
						dividerHrdAddr, dividerSrc, err := getNextHopHwAddr(localDividerIP)
						if err != nil {
							klog.Fatalf("error in getting dividernext hop hardware address: %v", err)
						}
						localDividerHosts = append(localDividerHosts, LocalDividerHost{IP: dividerSrc, Mac: dividerHrdAddr})
						gateway.LocalDividerHosts = localDividerHosts
						watcher.vpcGatewayMap[divider.Spec.Vni] = gateway
					}
				}
			}
		},
	})

	go watcher.gatewayconfigmapInformer.Run(watcher.quit)
	go watcher.subnetInformer.Run(watcher.quit)
	go watcher.dividerInformer.Run(watcher.quit)
	go watcher.vpcInformer.Run(watcher.quit)
	<-watcher.quit
}

func (watcher *KubeWatcher) SaveSubnet(payload []byte) {
	var subnet subnetv1.Subnet
	err := json.Unmarshal([]byte(payload), &subnet)
	subnet.ResourceVersion = ""
	subnet.Spec.Virtual = true
	if err != nil {
		log.Fatal(err)
	}
	if _, err = watcher.subnetInterface.Create(context.TODO(), &subnet, metav1.CreateOptions{}); err != nil {
		log.Fatal(err)
	}
}
