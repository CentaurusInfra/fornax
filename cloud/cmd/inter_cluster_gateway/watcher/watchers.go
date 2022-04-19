package watcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
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

type GatewayConfig struct {
	RemoteGatewayHosts []RemoteGatewayHost
	LocalDividerHosts  []LocalDividerHost
}

type RemoteGatewayHost struct {
	Remote net.IP
	IP     net.IP
	Mac    net.HardwareAddr
	Port   int
}

type LocalDividerHost struct {
	IP  net.IP
	Mac net.HardwareAddr
}

const (
	GenevePort    = 6081
	FirepowerPort = 2615
)

type NextHopAddr struct {
	LocalIP net.IP
	SrcIP   net.IP
	HrdAddr net.HardwareAddr
}

var (
	router           routing.Router
	handle           *pcap.Handle
	opts             gopacket.SerializeOptions
	buffer           gopacket.SerializeBuffer
	localHostIP      net.IP
	localGatewayHost string
)

type PacketWatcher struct {
	router routing.Router
	handle *pcap.Handle
	opts   gopacket.SerializeOptions
	buffer gopacket.SerializeBuffer
}

func NewPacketWatcher(router routing.Router, handle *pcap.Handle, buffer gopacket.SerializeBuffer) *PacketWatcher {
	return &PacketWatcher{
		router: router,
		handle: handle,
		buffer: buffer,
		opts: gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		},
	}
}

// GetNextHopHwAddr() finds out the next-hop hardware address if we want to send the packet to the destination IP
func (pw *PacketWatcher) GetNextHopHwAddr(destIP net.IP) (net.HardwareAddr, net.IP, error) {
	iface, gateway, src, err := pw.router.Route(destIP)
	if err != nil {
		return nil, nil, fmt.Errorf("error in getting icgw route : %v", err)
	}

	start := time.Now()
	arpDst := destIP
	if gateway != nil {
		arpDst = gateway
	}
	// Prepare the layers to send for an ARP request.
	eth := layers.Ethernet{
		SrcMAC:       iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(iface.HardwareAddr),
		SourceProtAddress: []byte(src),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
		DstProtAddress:    []byte(arpDst),
	}
	// Send a single ARP request packet since it is just a PoC work, may consider retry in the future
	if err := pw.Send(&eth, &arp); err != nil {
		return nil, nil, err
	}
	// Wait 3 seconds for an ARP reply.
	for {
		if time.Since(start) > time.Second*3 {
			return nil, nil, errors.New("timeout getting ARP reply")
		}
		data, _, err := pw.handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
			return nil, nil, err
		}
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
			arp := arpLayer.(*layers.ARP)
			if net.IP(arp.SourceProtAddress).Equal(net.IP(arpDst)) {
				return net.HardwareAddr(arp.SourceHwAddress), src, nil
			}
		}
	}
}

func (pw *PacketWatcher) Send(l ...gopacket.SerializableLayer) error {
	if err := gopacket.SerializeLayers(pw.buffer, pw.opts, l...); err != nil {
		return err
	}

	return pw.handle.WritePacketData(pw.buffer.Bytes())
}

type KubeWatcher struct {
	kubeconfig               *rest.Config
	quit                     chan struct{}
	subnetInformer           cache.SharedIndexInformer
	vpcInformer              cache.SharedIndexInformer
	dividerInformer          cache.SharedIndexInformer
	gatewayconfigmapInformer cache.SharedIndexInformer
	subnetMap                map[string]subnetv1.Subnet
	dividerMap               map[*net.IPNet]NextHopAddr
	vpcMap                   map[string]*vpcv1.Vpc
	vpcGatewayMap            map[string]GatewayConfig
	gatewayMap               map[string]string   //for neighbors
	remoteVpcIPMap           map[string][]string //for subnets
	subnetGatewayMap         map[*net.IPNet]NextHopAddr
	subnetInterface          subnetinterface.SubnetInterface
	gatewayName              string
	gatewayHostIP            string
	client                   *srv.Client
	packetWatcher            *PacketWatcher
}

func NewKubeWatcher(kubeConfig *rest.Config, client *srv.Client, packetWatcher *PacketWatcher, quit chan struct{}) (*KubeWatcher, error) {
	subMap := make(map[string]subnetv1.Subnet)
	if subclientset, err := subnetclientset.NewForConfig(kubeConfig); err == nil {
		if subList, err := subclientset.MizarV1().Subnets("default").List(context.TODO(), metav1.ListOptions{}); err == nil {
			for _, sub := range subList.Items {
				subMap[sub.Name] = sub
			}
		}
	} else {
		return nil, err
	}

	vpcMap := make(map[string]*vpcv1.Vpc)
	if vpcclientset, err := vpcclientset.NewForConfig(kubeConfig); err == nil {
		if vpcList, err := vpcclientset.MizarV1().Vpcs("default").List(context.TODO(), metav1.ListOptions{}); err == nil {
			for _, vpc := range vpcList.Items {
				vpcMap[vpc.Name] = &vpc
			}
		}
	} else {
		return nil, err
	}

	divMap := make(map[*net.IPNet]NextHopAddr)
	if divclientset, err := dividerclientset.NewForConfig(kubeConfig); err == nil {
		if divList, err := divclientset.MizarV1().Dividers("default").List(context.TODO(), metav1.ListOptions{}); err == nil {
			for _, div := range divList.Items {
				if key, val, err := getDividerMapEntryvpcMap(vpcMap, &div, packetWatcher); err == nil {
					divMap[key] = val
				} else {
					klog.Warningf("failed to generate divider map entry with the error %v", err)
				}
			}
		}
	} else {
		return nil, err
	}

	gatewayconfigmapclientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	gatewayMap := make(map[string]string)
	remoteVpcIPMap := make(map[string][]string)
	var subGatewayMap map[*net.IPNet]NextHopAddr
	var gatewayInfo *v1.ConfigMap
	if gatewayInfo, err = gatewayconfigmapclientset.CoreV1().ConfigMaps("default").Get(context.TODO(), constants.ClusterGatewayConfigMap, metav1.GetOptions{}); err == nil {
		if gateways := gatewayInfo.Data[constants.ClusterGatewayConfigMapGatewayNeighbors]; gateways != "" {
			gatewayArr := strings.Split(gateways, ",")
			for _, gatewayPair := range gatewayArr {
				gateway := strings.Split(gatewayPair, "=")
				gatewayMap[gateway[0]] = gateway[1]
			}
		}
		if remoteVpcIPs := gatewayInfo.Data[constants.ClusterGatewayConfigMapVpcGateways]; remoteVpcIPs != "" {
			remoteVpcIPMap = addRemoteVpcIPList(remoteVpcIPs)
		}
		klog.V(3).Infof("The current remoteVpcIPMap is %v", remoteVpcIPMap)
		if subGateways := gatewayInfo.Data[constants.ClusterGatewayConfigMapSubGateways]; subGateways != "" {
			subGatewayMap = getSubnetGatewayMap(subGateways, subMap, packetWatcher)
		}
	}

	klog.V(3).Infof("The current gateway is %v", gatewayMap)
	gatewayconfigmapSelector := fields.ParseSelectorOrDie("metadata.name=cluster-gateway-config")
	gatewayconfigmapLW := cache.NewListWatchFromClient(gatewayconfigmapclientset.CoreV1().RESTClient(), "configmaps", v1.NamespaceAll, gatewayconfigmapSelector)
	gatewayconfigmapInformer := cache.NewSharedIndexInformer(gatewayconfigmapLW, &v1.ConfigMap{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	subnetclientset, err := subnetclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	subnetSelector := fields.ParseSelectorOrDie("metadata.name!=net0")
	subnetLW := cache.NewListWatchFromClient(subnetclientset.MizarV1().RESTClient(), "subnets", v1.NamespaceAll, subnetSelector)
	subnetInformer := cache.NewSharedIndexInformer(subnetLW, &subnetv1.Subnet{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	vpcclientset, err := vpcclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	vpcSelector := fields.ParseSelectorOrDie("metadata.name!=vpc0")
	vpcLW := cache.NewListWatchFromClient(vpcclientset.MizarV1().RESTClient(), "vpcs", v1.NamespaceAll, vpcSelector)
	vpcInformer := cache.NewSharedIndexInformer(vpcLW, &vpcv1.Vpc{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	dividerclientset, err := dividerclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	dividerLW := cache.NewListWatchFromClient(dividerclientset.MizarV1().RESTClient(), "dividers", v1.NamespaceAll, fields.Everything())
	dividerInformer := cache.NewSharedIndexInformer(dividerLW, &dividerv1.Divider{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	return &KubeWatcher{
		kubeconfig:               kubeConfig,
		quit:                     quit,
		gatewayconfigmapInformer: gatewayconfigmapInformer,
		subnetInformer:           subnetInformer,
		vpcInformer:              vpcInformer,
		dividerInformer:          dividerInformer,
		subnetMap:                subMap,
		dividerMap:               divMap,
		vpcMap:                   vpcMap,
		vpcGatewayMap:            make(map[string]GatewayConfig),
		subnetGatewayMap:         subGatewayMap,
		gatewayMap:               gatewayMap,
		remoteVpcIPMap:           remoteVpcIPMap,
		subnetInterface:          subnetclientset.MizarV1().Subnets("default"),
		client:                   client,
		packetWatcher:            packetWatcher,
		gatewayName:              gatewayInfo.Data[constants.ClusterGatewayConfigMapGatewayName],
		gatewayHostIP:            gatewayInfo.Data[constants.ClusterGatewayConfigMapGatewayHostIP],
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
				if remoteVpcIPs := cm.Data[constants.ClusterGatewayConfigMapVpcGateways]; remoteVpcIPs != "" {
					watcher.remoteVpcIPMap = addRemoteVpcIPList(remoteVpcIPs)
				}
				klog.V(3).Infof("the remote vpc ip map is set to %v", watcher.remoteVpcIPMap)
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
			// Updated the remote vpc gateway ips
			if oldCm.Data[constants.ClusterGatewayConfigMapVpcGateways] != newCm.Data[constants.ClusterGatewayConfigMapVpcGateways] {
				if remoteVpcIPs := newCm.Data[constants.ClusterGatewayConfigMapVpcGateways]; remoteVpcIPs != "" {
					watcher.remoteVpcIPMap = addRemoteVpcIPList(remoteVpcIPs)
				} else {
					watcher.remoteVpcIPMap = make(map[string][]string)
				}
				klog.V(3).Infof("the remote vpc ip map is updated to to %v", watcher.remoteVpcIPMap)
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
			if oldCm.Data[constants.ClusterGatewayConfigMapSubGateways] != newCm.Data[constants.ClusterGatewayConfigMapSubGateways] {
				watcher.subnetGatewayMap = getSubnetGatewayMap(newCm.Data[constants.ClusterGatewayConfigMapSubGateways], watcher.subnetMap, watcher.packetWatcher)
			}
		},
	})
	watcher.subnetInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if subnet, ok := obj.(*subnetv1.Subnet); ok {
				klog.V(3).Infof("a new subnet %s is created", subnet.Name)
				if !subnet.Spec.Virtual {
					if remoteVpcIPList, ok := watcher.remoteVpcIPMap[subnet.Spec.Vpc]; ok {
						for _, remoteVpcIP := range remoteVpcIPList {
							klog.V(3).Infof("a new subnet %s is trying to sync to a remote vpc %s with the ip %s", subnet.Name, subnet.Spec.Vpc, remoteVpcIP)
							conn, client, ctx, cancel, err := watcher.client.Connect(remoteVpcIP)
							if err != nil {
								klog.Errorf("failed to sync a new subnet to %s with the error %v", remoteVpcIP, err)
							}
							defer conn.Close()
							defer cancel()
							request := &proto.CreateSubnetRequest{
								Name: subnet.Name, Namespace: subnet.Namespace, IP: subnet.Spec.IP, Status: subnet.Spec.Status,
								Prefix: subnet.Spec.Prefix, Vpc: subnet.Spec.Vpc, Vni: subnet.Spec.Vni, RemoteGateway: watcher.gatewayHostIP,
								Bouncers: int32(subnet.Spec.Bouncers)}
							returnMessage, err := client.CreateSubnet(ctx, request)
							klog.V(3).Infof("the returnMessage is %v", returnMessage)
							if err != nil {
								klog.Errorf("return from %s with the message %v with the error %v", remoteVpcIP, returnMessage, err)
							}
						}
					}
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if subnet, ok := obj.(*subnetv1.Subnet); ok {
				klog.V(3).Infof("a new subnet %s is deleted", subnet.Name)
				if !subnet.Spec.Virtual {
					if remoteVpcIPList, ok := watcher.remoteVpcIPMap[subnet.Spec.Vpc]; ok {
						for _, remoteVpcIP := range remoteVpcIPList {
							klog.V(3).Infof("a deleted subnet %s is trying to sync to a remote vpc %s with the ip %s", subnet.Name, subnet.Spec.Vpc, remoteVpcIP)
							conn, client, ctx, cancel, err := watcher.client.Connect(remoteVpcIP)
							if err != nil {
								klog.Errorf("failed to sync a new subnet to %s with the error %v", remoteVpcIP, err)
							}
							defer conn.Close()
							defer cancel()
							request := &proto.DeleteSubnetRequest{
								Name: subnet.Name, Namespace: subnet.Namespace}
							returnMessage, err := client.DeleteSubnet(ctx, request)
							klog.V(3).Infof("the returnMessage is %v", returnMessage)
							if err != nil {
								klog.Errorf("return from %s with the message %v with the error %v", remoteVpcIP, returnMessage, err)
							}
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
				watcher.vpcMap[vpc.Name] = vpc
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
				delete(watcher.vpcMap, vpc.Name)
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
				if key, val, err := getDividerMapEntryvpcMap(watcher.vpcMap, divider, watcher.packetWatcher); err == nil {
					watcher.dividerMap[key] = val
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

func (watcher *KubeWatcher) GetDividerMap() map[*net.IPNet]NextHopAddr {
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

func getDividerMapEntryvpcMap(vpcMap map[string]*vpcv1.Vpc, divider *dividerv1.Divider, pw *PacketWatcher) (*net.IPNet, NextHopAddr, error) {
	if _, existed := vpcMap[divider.Spec.Vpc]; existed {
		vpc := vpcMap[divider.Spec.Vpc]
		if _, cidr, err := net.ParseCIDR(fmt.Sprintf("%s/%s", vpc.Spec.IP, vpc.Spec.Prefix)); err == nil {
			localDividerIP := net.ParseIP(divider.Spec.IP)
			if localDividerIP == nil {
				klog.Errorf("Invalid local divider IP: %v", divider.Spec.IP)
				return nil, NextHopAddr{}, err
			}
			localDividerIP = localDividerIP.To4()
			dividerHrdAddr, dividerSrc, err := pw.GetNextHopHwAddr(localDividerIP)
			klog.V(3).Infof("The divider ip %v and mac %v", divider.Spec.IP, divider.Spec.Mac)
			klog.V(3).Infof("The divider src ip %v and mac %v", dividerSrc, dividerHrdAddr)
			if err != nil {
				klog.Errorf("error in getting dividernext hop hardware address: %v", err)
				return nil, NextHopAddr{}, err
			}
			return cidr, NextHopAddr{LocalIP: localDividerIP, SrcIP: dividerSrc, HrdAddr: dividerHrdAddr}, nil
		}
	}
	return nil, NextHopAddr{}, nil
}

func getSubnetGatewayMap(subGateways string, subnetMap map[string]subnetv1.Subnet, packetWatcher *PacketWatcher) map[*net.IPNet]NextHopAddr {
	subnetGatewayMap := make(map[*net.IPNet]NextHopAddr)
	subGatewayArr := strings.Split(subGateways, ",")
	for _, subGateway := range subGatewayArr {
		subGatewayPair := strings.Split(subGateway, "=")
		if sub, ok := subnetMap[subGatewayPair[0]]; ok {
			if _, cidr, err := net.ParseCIDR(fmt.Sprintf("%s/%s", sub.Spec.IP, sub.Spec.Prefix)); err == nil {
				remoteIcgwIP := net.ParseIP(subGatewayPair[1])
				if remoteIcgwIP == nil {
					klog.Errorf("Invalid remote gateway host  IP: %v", subGatewayPair[1])
				}
				remoteIcgwIP = remoteIcgwIP.To4()
				icgwHrdAddr, icgwSrc, err := packetWatcher.GetNextHopHwAddr(remoteIcgwIP)
				if err != nil {
					klog.Errorf("error in getting divider next hop hardware address: %v", err)
				} else {
					subnetGatewayMap[cidr] = NextHopAddr{LocalIP: remoteIcgwIP, SrcIP: icgwSrc, HrdAddr: icgwHrdAddr}
					klog.V(3).Infof("The remote gateway ip %v and mac %v", icgwSrc, icgwHrdAddr)
				}
			}
		}
	}
	return subnetGatewayMap
}

func addRemoteVpcIPList(remoteVpcIPs string) map[string][]string {
	remoteVpcIPMap := make(map[string][]string)
	remoteVpcIPArr := strings.Split(remoteVpcIPs, ",")
	for _, vpcIPPair := range remoteVpcIPArr {
		vpcIP := strings.Split(vpcIPPair, "=")
		if IPList, ok := remoteVpcIPMap[vpcIP[0]]; ok {
			if !contains(IPList, vpcIP[1]) {
				IPList = append(IPList, vpcIP[1])
			}
			remoteVpcIPMap[vpcIP[0]] = IPList
		} else {
			remoteVpcIPMap[vpcIP[0]] = []string{vpcIP[1]}
		}
	}
	return remoteVpcIPMap
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
