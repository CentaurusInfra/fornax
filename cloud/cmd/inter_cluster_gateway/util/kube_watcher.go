package util

import (
	"log"
	"net"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	dividerclientset "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/divider/client/clientset/versioned"
	dividerv1 "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/divider/v1"
	subnetclientset "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/subnet/client/clientset/versioned"
	subnetv1 "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/subnet/v1"
	vpcclientset "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/vpc/client/clientset/versioned"
	vpcv1 "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/vpc/v1"
)

type KubeWatcher struct {
	kubeconfig      *rest.Config
	quit            chan struct{}
	subnetInformer  cache.SharedIndexInformer
	vpcInformer     cache.SharedIndexInformer
	dividerInformer cache.SharedIndexInformer
	subnetMap       map[string]*subnetv1.Subnet
	dividerMap      map[string]*dividerv1.Divider
	vpcMap          map[string]*vpcv1.Vpc
	vpcGatewayMap   map[string]GatewayConfig
}

func NewKubeWatcher(kubeconfig *rest.Config, quit chan struct{}) (*KubeWatcher, error) {
	subnetclientset, err := subnetclientset.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	subnetLW := cache.NewListWatchFromClient(subnetclientset.MizarV1().RESTClient(), "subnets", v1.NamespaceAll, fields.Everything())
	subnetInformer := cache.NewSharedIndexInformer(subnetLW, &subnetv1.Subnet{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	vpcclientset, err := vpcclientset.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	vpcLW := cache.NewListWatchFromClient(vpcclientset.MizarV1().RESTClient(), "vpcs", v1.NamespaceAll, fields.Everything())
	vpcInformer := cache.NewSharedIndexInformer(vpcLW, &vpcv1.Vpc{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	dividerclientset, err := dividerclientset.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	dividerLW := cache.NewListWatchFromClient(dividerclientset.MizarV1().RESTClient(), "dividers", v1.NamespaceAll, fields.Everything())
	dividerInformer := cache.NewSharedIndexInformer(dividerLW, &dividerv1.Divider{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	return &KubeWatcher{
		kubeconfig:      kubeconfig,
		quit:            quit,
		subnetInformer:  subnetInformer,
		vpcInformer:     vpcInformer,
		dividerInformer: dividerInformer,
		subnetMap:       make(map[string]*subnetv1.Subnet),
		dividerMap:      make(map[string]*dividerv1.Divider),
		vpcMap:          make(map[string]*vpcv1.Vpc),
		vpcGatewayMap:   make(map[string]GatewayConfig),
	}, nil
}

func (watcher *KubeWatcher) Run() {
	watcher.subnetInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if subnet, ok := obj.(*subnetv1.Subnet); ok {
				klog.V(3).Infof("A new subnet %s is trying to sync to %s", subnet.Name, subnet.Spec.RemoteGateways)
				if _, existed := watcher.subnetMap[subnet.Name]; !existed {
					watcher.subnetMap[subnet.Name] = subnet
					if !subnet.Spec.Virtual {
						subnet.Spec.RemoteGateways = []string{localGatewayHost}
						klog.V(3).Infof("The remote gateways are set to %v", subnet.Spec.RemoteGateways)
						if vpcGateway, existed := watcher.vpcGatewayMap[subnet.Spec.Vni]; existed {
							for _, remoteGatewayHost := range vpcGateway.RemoteGatewayHosts {
								if err := syncSubnet(subnet, remoteGatewayHost.Remote, remoteGatewayHost.IP, remoteGatewayHost.Mac, remoteGatewayHost.Port, int(GenevePort)); err != nil {
									klog.Fatalf("Error in synchronizing subnet %v: %v", subnet, err)
								}
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
				klog.V(3).Infof("A new vpc %s is trying to sync to %s", vpc.Name, vpc.Spec.RemoteGateways)
				if _, existed := watcher.vpcMap[vpc.Spec.Vni]; !existed {
					watcher.vpcMap[vpc.Spec.Vni] = vpc
					if len(vpc.Spec.RemoteGateways) > 0 {
						var remoteGatewayHosts []RemoteGatewayHost
						for _, host := range vpc.Spec.RemoteGateways {
							hosts := strings.Split(host, ":")
							klog.V(3).Infof("The gateway host ip is %s and the port is %s", hosts[0], hosts[1])
							remoteIcgwIP := net.ParseIP(hosts[0])
							if remoteIcgwIP == nil {
								klog.Fatalf("Invalid remote ICGW IP: %v", hosts[0])
							}
							remoteIcgwIP = remoteIcgwIP.To4()
							icgwHrdAddr, icgwSrc, err := getNextHopHwAddr(remoteIcgwIP)
							if err != nil {
								klog.Fatalf("error in getting divider next hop hardware address: %v", err)
							}
							port, err := strconv.Atoi(hosts[1])
							if err != nil {
								klog.Fatalf("error in getting remote gateway port: %s", hosts[1])
							}
							remoteGatewayHosts = append(remoteGatewayHosts, RemoteGatewayHost{Remote: remoteIcgwIP, IP: icgwSrc, Port: port, Mac: icgwHrdAddr})
							klog.V(3).Infof("The remote gateway hosts are set to ", remoteGatewayHosts)
						}
						watcher.vpcGatewayMap[vpc.Spec.Vni] = GatewayConfig{RemoteGatewayHosts: remoteGatewayHosts}
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
					}
				}
			}
		},
	})

	go watcher.subnetInformer.Run(watcher.quit)
	go watcher.dividerInformer.Run(watcher.quit)
	go watcher.vpcInformer.Run(watcher.quit)
	<-watcher.quit
}

func GetLocalHostIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
