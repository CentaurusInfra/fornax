package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/kubeedge/kubeedge/cloud/cmd/config"
	"github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/srv"
	"github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/watcher"
)

var (
	kubeconfig    string
	genevePort    layers.UDPPort
	firepowerPort layers.UDPPort
	logLevel      string
	grpcPort      int
	grpcTimeout   int

	router routing.Router
	handle *pcap.Handle
	buffer gopacket.SerializeBuffer
)

// initialize the commandline options
func InitFlag() {
	flag.StringVar(&kubeconfig, "kubeconfig", "/etc/kubernetes/admin.conf", "the kubeconfig file path to access kube apiserver")

	config, err := config.NewGatewayConfiguration("gateway_config.json")
	if err != nil {
		panic(fmt.Errorf("error setting gateway agent configuration: %v", err))
	}

	logLevel = config.LogLevel
	local := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	klog.InitFlags(local)
	err = local.Set("v", logLevel)
	if err != nil {
		fmt.Printf("error setting klog flags: %v", err)
		os.Exit(1)
	}

	firepowerPort = layers.UDPPort((config.GatewayPort))
	genevePort = layers.UDPPort((config.GenevePort))
	grpcPort = config.GrpcPort
	grpcTimeout = config.GrpcTimeout
}

// Initialize the config vuales to use in packet forwarding
func init() {
	var err error
	InitFlag()

	buffer = gopacket.NewSerializeBuffer()

	// refer to https://www.youtube.com/watch?v=APDnbmTKjgM for a nice talk about what these values are
	handle, err = pcap.OpenLive("eth0", int32(65535), false, -1*time.Second)

	if err != nil {
		klog.Fatalf("error in initializing handle: %v", err)
	}

	router, err = routing.New()
	if err != nil {
		klog.Fatalf("error in initilization of router : %v", err)
	}
}

func main() {
	defer handle.Close()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	go srv.RunGrpcServer(config, grpcPort)

	quit := make(chan struct{})
	defer close(quit)
	packetWatcher := watcher.NewPacketWatcher(router, handle, buffer)
	kubeWatcher, err := watcher.NewKubeWatcher(config, srv.NewClient(grpcPort, grpcTimeout), packetWatcher, quit)
	if err != nil {
		panic(err)
	}
	go kubeWatcher.Run()

	geneveFilter := fmt.Sprintf("port %d", genevePort)
	if err := handle.SetBPFFilter(geneveFilter); err != nil {
		klog.Fatalf("Error in setting up BPF filter: %v", err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		if err := processPacket(&packet, kubeWatcher.GetSubnetGatewayMap(), kubeWatcher.GetDividerMap(), packetWatcher); err != nil {
			klog.Errorf("Failed to process packet: %v, packet: %v", err, packet)
		} else {
			klog.V(3).Infof("Successfully processed packet")
			klog.V(5).Infof("packet processed: %v", packet)
		}
	}
}

func processPacket(p *gopacket.Packet, gatewayMap map[*net.IPNet]watcher.NextHopAddr, dividerMap map[*net.IPNet]watcher.NextHopAddr, packetWatcher *watcher.PacketWatcher) error {
	packet := *p

	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	ethernetFrame := ethernetLayer.(*layers.Ethernet)
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	ipPacket, _ := ipLayer.(*layers.IPv4)
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	udpPacket := udpLayer.(*layers.UDP)
	ethernetFrameCopy := *ethernetFrame
	ipPacketCopy := *ipPacket
	udpPacketCopy := *udpPacket
	var internalDstIP net.IP
	if len(packet.Layers()) >= 5 {
		internalIPLayer := packet.Layers()[5]
		if internalIPLayer.LayerType() == layers.LayerTypeIPv4 {
			internalIPPacket, _ := internalIPLayer.(*layers.IPv4)
			internalDstIP = internalIPPacket.DstIP
		}
	}

	if internalDstIP == nil {
		klog.Errorf("failed to get internal dst ip from the packet %v", packet)
	}

	switch {
	// when receiving a geneve packet, note that only local clusters send geneve packets to the Inter-Cluster Gateway
	case udpPacketCopy.DstPort == genevePort:

		remoteIcgwIP, icgwSrc, icgwHrdAddr := getRemoteGateway(internalDstIP, gatewayMap)
		ethernetFrameCopy.SrcMAC = ethernetFrame.DstMAC
		ethernetFrameCopy.DstMAC = icgwHrdAddr

		ipPacketCopy.SrcIP = icgwSrc
		ipPacketCopy.DstIP = remoteIcgwIP

		// change the port so the receiving side mizar will not catch it, as the mizar XDP only catches geneve packets.
		// so the packet will be processed by the inter-cluster gateway in the user space.
		udpPacketCopy.SrcPort = genevePort
		udpPacketCopy.DstPort = firepowerPort

		klog.V(3).Infof("Forwarding packet to remote icgw %v", remoteIcgwIP)

	// it is a packet from the remote Inter-Cluster gateway
	case udpPacketCopy.SrcPort == genevePort && udpPacketCopy.DstPort == firepowerPort:
		localDividerIP, dividerSrc, dividerHrdAddr := getDivider(internalDstIP, dividerMap)
		ethernetFrameCopy.SrcMAC = ethernetFrame.DstMAC
		ethernetFrameCopy.DstMAC = dividerHrdAddr

		ipPacketCopy.SrcIP = dividerSrc
		ipPacketCopy.DstIP = localDividerIP

		// change the port packet becomes a geneve packet again, so it will be caught by the mizar XDP in the divider.
		udpPacketCopy.SrcPort = firepowerPort
		udpPacketCopy.DstPort = genevePort

		klog.V(3).Infof("Forwarding packet to local divider %v", localDividerIP)

	default:
		return fmt.Errorf("unsupported packages")
	}

	buffer = gopacket.NewSerializeBuffer()

	if err := udpPacketCopy.SetNetworkLayerForChecksum(&ipPacketCopy); err != nil {
		return fmt.Errorf("error in setup network layer checksum: %v", err)
	}

	return packetWatcher.Send(&ethernetFrameCopy, &ipPacketCopy, &udpPacketCopy, gopacket.Payload(udpPacketCopy.Payload))
}

func getRemoteGateway(dstIP net.IP, gatewayMap map[*net.IPNet]watcher.NextHopAddr) (net.IP, net.IP, net.HardwareAddr) {
	for cidr, addr := range gatewayMap {
		if cidr.Contains(dstIP) {
			return addr.LocalIP, addr.SrcIP, addr.HrdAddr
		}
	}
	return nil, nil, nil
}

func getDivider(dstIP net.IP, dividerMap map[*net.IPNet]watcher.NextHopAddr) (net.IP, net.IP, net.HardwareAddr) {
	for cidr, divider := range dividerMap {
		if cidr.Contains(dstIP) {
			return divider.LocalIP, divider.SrcIP, divider.HrdAddr
		}
	}
	return nil, nil, nil
}
