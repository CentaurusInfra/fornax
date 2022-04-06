package util

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/kubeedge/kubeedge/common/constants"
)

type KubeWatcher struct {
	kubeconfig               *rest.Config
	quit                     chan struct{}
	subnetInformer           cache.SharedIndexInformer
	vpcInformer              cache.SharedIndexInformer
	dividerInformer          cache.SharedIndexInformer
	gatewayconfigmapInformer cache.SharedIndexInformer
	subnetMap                map[string]*subnetv1.Subnet
	dividerMap               map[*net.IPNet][]NextHopAddr
	vpcMap                   map[string]*vpcv1.Vpc
	vpcGatewayMap            map[string]GatewayConfig
	gatewayMap               map[string]string
	subnetGatewayMap         map[*net.IPNet]NextHopAddr
	subnetInterface          subnetinterface.SubnetInterface
	gatewayName              string
	gatewayHostIP            string
	client                   *srv.Client
}

type NextHopAddr struct {
	LocalIP net.IP
	SrcIP   net.IP
	HrdAddr net.HardwareAddr
}

func NewKubeWatcher(kubeconfig *rest.Config, client *srv.Client, quit chan struct{}) (*KubeWatcher, error) {
	gatewayconfigmapclientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	gatewayMap := make(map[string]string)
	if gatewayInfo, err := gatewayconfigmapclientset.CoreV1().ConfigMaps("default").Get(context.TODO(), "cluster-gateway-config", metav1.GetOptions{}); err == nil {
		if gateways := gatewayInfo.Data["vpc_gateways"]; gateways != "" {
			gatewayArr := strings.Split(gateways, ",")
			for _, gatewayPair := range gatewayArr {
				gateway := strings.Split(gatewayPair, "=")
				gatewayMap[gateway[0]] = gateway[1]
			}
		}
	}
	klog.V(3).Infof("The current gateway is %v", gatewayMap)
	gatewayconfigmapSelector := fields.ParseSelectorOrDie("metadata.name=cluster-gateway-config")
	gatewayconfigmapLW := cache.NewListWatchFromClient(gatewayconfigmapclientset.CoreV1().RESTClient(), "configmaps", v1.NamespaceAll, gatewayconfigmapSelector)
	gatewayconfigmapInformer := cache.NewSharedIndexInformer(gatewayconfigmapLW, &v1.ConfigMap{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	subnetclientset, err := subnetclientset.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	subnetSelector := fields.ParseSelectorOrDie("metadata.name!=net0")
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
		dividerMap:               make(map[*net.IPNet][]NextHopAddr),
		vpcMap:                   make(map[string]*vpcv1.Vpc),
		vpcGatewayMap:            make(map[string]GatewayConfig),
		subnetGatewayMap:         make(map[*net.IPNet]NextHopAddr),
		gatewayMap:               gatewayMap,
		subnetInterface:          subnetclientset.MizarV1().Subnets("default"),
		client:                   client,
	}, nil
}

func (watcher *KubeWatcher) Run() {
	watcher.gatewayconfigmapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if cm, ok := obj.(*v1.ConfigMap); ok {
				watcher.gatewayName = cm.Data[constants.ClusterGatewayConfigMapGatewayName]
				watcher.gatewayHostIP = cm.Data[constants.ClusterGatewayConfigMapGatewayHostIP]
				gatewayArr := strings.Split(cm.Data[constants.ClusterGatewayConfigMapGatewayNeighbors], ",")
				for _, gatewayPair := range gatewayArr {
					gateway := strings.Split(gatewayPair, "=")
					watcher.gatewayMap[gateway[0]] = gateway[1]
				}
				klog.V(3).Infof("the gateway map is set to %v", watcher.gatewayMap)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			oldCm, newCm := old.(*v1.ConfigMap), new.(*v1.ConfigMap)
			// Updated the current gateway neighbors
			if oldCm.Data[constants.ClusterGatewayConfigMapGatewayNeighbors] != newCm.Data[constants.ClusterGatewayConfigMapGatewayNeighbors] {
				watcher.gatewayMap = make(map[string]string)
				gatewayArr := strings.Split(newCm.Data[constants.ClusterGatewayConfigMapGatewayNeighbors], ",")
				for _, gatewayPair := range gatewayArr {
					gateway := strings.Split(gatewayPair, "=")
					watcher.gatewayMap[gateway[0]] = gateway[1]
				}
				klog.V(3).Infof("the gateway map is updated to %v", watcher.gatewayMap)
			}
			// Updated the vpc gateway info
			klog.V(3).Infof("a new vpc is created/deleted in a neighbor and the vpc gateway pair is updated from %s to %s",
				oldCm.Data[constants.ClusterGatewayConfigMapVpcGateways], newCm.Data[constants.ClusterGatewayConfigMapVpcGateways])
			compare, diff := compareAndDiffVpcGateways(newCm.Data[constants.ClusterGatewayConfigMapVpcGateways],
				oldCm.Data[constants.ClusterGatewayConfigMapVpcGateways])
			if compare != 0 {
				for _, gatewayHostIP := range watcher.gatewayMap {
					conn, client, ctx, cancel, err := watcher.client.Connect(gatewayHostIP)
					if err != nil {
						klog.Errorf("failed to sync a deleted subnet to %s with the error %v", gatewayHostIP, err)
					}
					defer conn.Close()
					defer cancel()
					vpcGateway := strings.Split(diff, "=")
					var returnMessage *proto.Response
					if compare == 1 && vpcGateway[1] != gatewayHostIP {
						request := &proto.CreateVpcGatewayRequest{
							Name: vpcGateway[0], Namespace: "default", GatewayHostIP: vpcGateway[1]}
						returnMessage, err = client.CreateVpcGateway(ctx, request)
					} else if compare == -1 && vpcGateway[1] != gatewayHostIP {
						request := &proto.DeleteVpcGatewayRequest{
							Name: vpcGateway[0], Namespace: "default", GatewayHostIP: vpcGateway[1]}
						returnMessage, err = client.DeleteVpcGateway(ctx, request)
					}
					klog.V(3).Infof("The returnMessage is %v", returnMessage)
					if err != nil {
						klog.Errorf("return from %s with the message %v with the error %v", gatewayHostIP, returnMessage, err)
					}
				}
			}
		},
	})
	watcher.subnetInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if subnet, ok := obj.(*subnetv1.Subnet); ok {
				klog.V(3).Infof("a new subnet %s is created", subnet.Name)
				if subnet.Spec.Virtual {
					if len(subnet.Spec.RemoteGateways) == 0 {
						return
					}
					if _, cidr, err := net.ParseCIDR(fmt.Sprintf("%s/%s", subnet.Spec.IP, subnet.Spec.Prefix)); err == nil {
						remoteIcgwIP := net.ParseIP(subnet.Spec.RemoteGateways[0])
						if remoteIcgwIP == nil {
							klog.Errorf("Invalid remote gateway host  IP: %v", subnet.Spec.RemoteGateways[0])
						}
						remoteIcgwIP = remoteIcgwIP.To4()
						icgwHrdAddr, icgwSrc, err := GetNextHopHwAddr(remoteIcgwIP)
						if err != nil {
							klog.Errorf("error in getting divider next hop hardware address: %v", err)
						} else {
							watcher.subnetGatewayMap[cidr] = NextHopAddr{LocalIP: remoteIcgwIP, SrcIP: icgwSrc, HrdAddr: icgwHrdAddr}
							klog.V(3).Infof("The remote gateway ip %v and mac %v", icgwSrc, icgwHrdAddr)
						}
					}
				} else {
					for gatewayName, gatewayHostIP := range watcher.gatewayMap {
						klog.V(3).Infof("a new subnet %s is trying to sync to %s with the ip %s", subnet.Name, gatewayName, gatewayHostIP)
						conn, client, ctx, cancel, err := watcher.client.Connect(gatewayHostIP)
						if err != nil {
							klog.Errorf("failed to sync a new subnet to %s with the error %v", gatewayHostIP, err)
						}
						defer conn.Close()
						defer cancel()
						request := &proto.CreateSubnetRequest{
							Name: subnet.Name, Namespace: subnet.Namespace, IP: subnet.Spec.IP, Status: subnet.Spec.Status,
							Prefix: subnet.Spec.Prefix, Vpc: subnet.Spec.Vpc, Vni: subnet.Spec.Vni, RemoteGateway: watcher.gatewayHostIP}
						returnMessage, err := client.CreateSubnet(ctx, request)
						klog.V(3).Infof("the returnMessage is %v", returnMessage)
						if err != nil {
							klog.Errorf("return from %s with the message %v with the error %v", gatewayHostIP, returnMessage, err)
						}
					}
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if subnet, ok := obj.(*subnetv1.Subnet); ok {
				klog.V(3).Infof("a new subnet %s is deleted", subnet.Name)
				if subnet.Spec.Virtual {
					if len(subnet.Spec.RemoteGateways) == 0 {
						return
					}
					if _, cidr, err := net.ParseCIDR(fmt.Sprintf("%s/%s", subnet.Spec.IP, subnet.Spec.Prefix)); err == nil {
						delete(watcher.subnetGatewayMap, cidr)
					}
				} else {
					for gatewayName, gatewayHostIP := range watcher.gatewayMap {
						klog.V(3).Infof("a deleted subnet %s is trying to sync to %s with the ip %s", subnet.Name, gatewayName, gatewayHostIP)
						conn, client, ctx, cancel, err := watcher.client.Connect(gatewayHostIP)
						if err != nil {
							klog.Errorf("failed to sync a deleted subnet to %s with the error %v", gatewayHostIP, err)
						}
						defer conn.Close()
						defer cancel()
						request := &proto.DeleteSubnetRequest{
							Name: subnet.Name, Namespace: subnet.Namespace}
						returnMessage, err := client.DeleteSubnet(ctx, request)
						klog.V(3).Infof("The returnMessage is %v", returnMessage)
						if err != nil {
							klog.Errorf("return from %s with the message %v with the error %v", gatewayHostIP, returnMessage, err)
						}
					}
				}
			}
		},
	})

	watcher.vpcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if vpc, ok := obj.(*vpcv1.Vpc); ok {
				klog.V(3).Infof("a new vpc %s is created", vpc.Name)
				watcher.vpcMap[vpc.Spec.Vni] = vpc
				for gatewayName, gatewayHostIP := range watcher.gatewayMap {
					klog.V(3).Infof("a new vpc %s is trying to sync to %s with the ip %s", vpc.Name, gatewayName, gatewayHostIP)
					conn, client, ctx, cancel, err := watcher.client.Connect(gatewayHostIP)
					if err != nil {
						klog.Errorf("failed to sync a new vpc to %s with the error %v", gatewayHostIP, err)
					}
					defer conn.Close()
					defer cancel()
					request := &proto.CreateVpcGatewayRequest{Name: vpc.Name, Namespace: vpc.Namespace, GatewayName: watcher.gatewayName, GatewayHostIP: watcher.gatewayHostIP}
					returnMessage, err := client.CreateVpcGateway(ctx, request)
					klog.V(3).Infof("the returnMessage is %v", returnMessage)
					if err != nil {
						klog.Errorf("return from %s with the message %v with the error %v", gatewayHostIP, returnMessage, err)
					}
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if vpc, ok := obj.(*vpcv1.Vpc); ok {
				klog.V(3).Infof("a existing vpc %s is deletedc", vpc.Name)
				delete(watcher.vpcMap, vpc.Spec.Vni)
				for gatewayName, gatewayHostIP := range watcher.gatewayMap {
					klog.V(3).Infof("a deleted vpc %s is trying to sync to %s with the ip %s", vpc.Name, gatewayName, gatewayHostIP)
					conn, client, ctx, cancel, err := watcher.client.Connect(gatewayHostIP)
					if err != nil {
						klog.Errorf("failed to sync a deleted vpc to %s with the error %v", gatewayHostIP, err)
					}
					defer conn.Close()
					defer cancel()
					request := &proto.DeleteVpcGatewayRequest{Name: vpc.Name, Namespace: vpc.Namespace, GatewayName: watcher.gatewayName, GatewayHostIP: watcher.gatewayHostIP}
					returnMessage, err := client.DeleteVpcGateway(ctx, request)
					klog.V(3).Infof("the returnMessage is %v", returnMessage)
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
				if _, existed := watcher.vpcMap[divider.Spec.Vni]; existed {
					vpc := watcher.vpcMap[divider.Spec.Vni]
					if _, cidr, err := net.ParseCIDR(fmt.Sprintf("%s/%s", vpc.Spec.IP, vpc.Spec.Prefix)); err == nil {
						if _, existed := watcher.dividerMap[cidr]; !existed {
							watcher.dividerMap[cidr] = make([]NextHopAddr, 0)
						}
						localDividerIP := net.ParseIP(divider.Spec.IP)
						if localDividerIP == nil {
							klog.Errorf("Invalid local divider IP: %v", divider.Spec.IP)
						}
						localDividerIP = localDividerIP.To4()
						dividerHrdAddr, dividerSrc, err := GetNextHopHwAddr(localDividerIP)
						klog.V(3).Infof("The divider ip %v and mac %v", divider.Spec.IP, divider.Spec.Mac)
						klog.V(3).Infof("The divider src ip %v and mac %v", dividerSrc, dividerHrdAddr)
						if err != nil {
							klog.Errorf("error in getting dividernext hop hardware address: %v", err)
						} else {
							watcher.dividerMap[cidr] = append(watcher.dividerMap[cidr], NextHopAddr{LocalIP: localDividerIP, SrcIP: dividerSrc, HrdAddr: dividerHrdAddr})
						}
					}
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if divider, ok := obj.(*dividerv1.Divider); ok {
				klog.V(3).Infof("A divider %s is deleted", divider.Name)
				if _, existed := watcher.vpcMap[divider.Spec.Vni]; !existed {
					vpc := watcher.vpcMap[divider.Spec.Vni]
					if _, cidr, err := net.ParseCIDR(fmt.Sprintf("%s/%s", vpc.Spec.IP, vpc.Spec.Prefix)); err == nil {
						delete(watcher.dividerMap, cidr)
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

func (watcher *KubeWatcher) GetSubnetGatewayMap() map[*net.IPNet]NextHopAddr {
	return watcher.subnetGatewayMap
}

func (watcher *KubeWatcher) GetDividerMap() map[*net.IPNet][]NextHopAddr {
	return watcher.dividerMap
}

func compareAndDiffVpcGateways(str1, str2 string) (int, string) {
	compareResult := 0
	diffResult := ""
	if len(str1) < len(str2) {
		compareResult = -1
		_, diffResult = compareAndDiffVpcGateways(str2, str1)
	} else if len(str1) > len(str2) {
		compareResult = 1
		str2Map := make(map[string]bool)
		str2Arr := strings.Split(str2, ",")
		for _, key := range str2Arr {
			str2Map[key] = true
		}
		str1Arr := strings.Split(str1, ",")
		for _, key := range str1Arr {
			if _, ok := str2Map[key]; !ok {
				return compareResult, key
			}
		}
	}
	return compareResult, diffResult
}
