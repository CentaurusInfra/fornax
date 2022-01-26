package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
	"github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/config"
	"k8s.io/klog/v2"
)

var (
	genevePort    layers.UDPPort
	firepowerPort layers.UDPPort
	logLevel      string

	remoteIcgwIPstr   string
	localDividerIPstr string

	localDividerIP net.IP
	remoteIcgwIP   net.IP
	router         routing.Router

	icgwHrdAddr    net.HardwareAddr
	dividerHrdAddr net.HardwareAddr

	icgwSrc    net.IP
	dividerSrc net.IP

	handle *pcap.Handle
	opts   gopacket.SerializeOptions
	buffer gopacket.SerializeBuffer
)

// initialize the commandline options
func InitFlag() {
	config, err := config.NewGatewayConfiguration("gateway_config.json")
	if err != nil {
		fmt.Printf("error setting gateway agent configuration: %v", err)
		os.Exit(1)
	}

	logLevel = string(config.LogLevel)
	local := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	klog.InitFlags(local)
	err = local.Set("v", logLevel)
	if err != nil {
		fmt.Printf("error setting klog flags: %v", err)
		os.Exit(1)
	}

	//There shall be multiple remote gateways. However, the current implementation only works for one gateway
	remoteIcgwIPstr = config.RemoteGateways[0].RemoteGatewayIP
	firepowerPort = layers.UDPPort((config.RemoteGateways[0].RemoteGatewayPort))
	localDividerIPstr = config.LocalDividerIP
	genevePort = layers.UDPPort((config.GenevePort))

	remoteIcgwIP = net.ParseIP(remoteIcgwIPstr)
	if remoteIcgwIP == nil {
		klog.Fatalf("Invalid remote ICGW IP: %v", remoteIcgwIPstr)
	}
	remoteIcgwIP = remoteIcgwIP.To4()

	localDividerIP = net.ParseIP(localDividerIPstr)
	if localDividerIP == nil {
		klog.Fatalf("Invalid local divider IP: %v", localDividerIPstr)
	}
	localDividerIP = localDividerIP.To4()
}

// Initialize the config vuales to use in packet forwarding
func init() {
	var err error
	InitFlag()

	opts = gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

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

	icgwHrdAddr, icgwSrc, err = getNextHopHwAddr(remoteIcgwIP)
	if err != nil {
		klog.Fatalf("error in getting divider next hop hardware address: %v", err)
	}

	dividerHrdAddr, dividerSrc, err = getNextHopHwAddr(localDividerIP)
	if err != nil {
		klog.Fatalf("error in getting dividernext hop hardware address: %v", err)
	}
}

func main() {
	defer handle.Close()

	version := pcap.Version()
	fmt.Println(version)

	geneveFilter := fmt.Sprintf("port %d", genevePort)
	if err := handle.SetBPFFilter(geneveFilter); err != nil {
		klog.Fatalf("Error in setting up BPF filter: %v", err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		if err := processPacket(&packet); err != nil {
			klog.Errorf("Failed to process packet: %v, packet: %v", err, packet)
		} else {
			klog.V(3).Infof("Successfully processed packet")
			klog.V(5).Infof("packet processed: %v", packet)
		}
	}
}

func processPacket(p *gopacket.Packet) error {
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

	switch {
	// when receiving a geneve packet, note that only local clusters send geneve packets to the Inter-Cluster Gateway
	case udpPacketCopy.DstPort == genevePort:
		ethernetFrameCopy.SrcMAC = ethernetFrame.DstMAC
		ethernetFrameCopy.DstMAC = icgwHrdAddr

		ipPacketCopy.SrcIP = icgwSrc
		ipPacketCopy.DstIP = remoteIcgwIP

		// change the port so the receiving side mizar will not catch it, as the mizar XDP only catches geneve packets.
		// so the packet will be processed by the inter-cluster gateway in the user space.
		udpPacketCopy.SrcPort = genevePort
		udpPacketCopy.DstPort = firepowerPort

		klog.V(3).Infof("Forwarding packet to remote icgw %v", remoteIcgwIPstr)

	// it is a packet from the remote Inter-Cluster gateway
	case udpPacketCopy.SrcPort == genevePort && udpPacketCopy.DstPort == firepowerPort:
		ethernetFrameCopy.SrcMAC = ethernetFrame.DstMAC
		ethernetFrameCopy.DstMAC = dividerHrdAddr

		ipPacketCopy.SrcIP = dividerSrc
		ipPacketCopy.DstIP = localDividerIP

		// change the port packet becomes a geneve packet again, so it will be caught by the mizar XDP in the divider.
		udpPacketCopy.SrcPort = firepowerPort
		udpPacketCopy.DstPort = genevePort

		klog.V(3).Infof("Forwarding packet to local divider %v", localDividerIPstr)

	default:
		return fmt.Errorf("Unsupported packages")
	}

	buffer = gopacket.NewSerializeBuffer()

	if err := udpPacketCopy.SetNetworkLayerForChecksum(&ipPacketCopy); err != nil {
		return fmt.Errorf("Error in setup network layer checksum: %v", err)
	}

	return send(&ethernetFrameCopy, &ipPacketCopy, &udpPacketCopy, gopacket.Payload(udpPacketCopy.Payload))
}

// getNextHopHwAddr() finds out the next-hop hardware address if we want to send the packet to the destination IP
func getNextHopHwAddr(destIP net.IP) (net.HardwareAddr, net.IP, error) {
	iface, gateway, src, err := router.Route(destIP)
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
	if err := send(&eth, &arp); err != nil {
		return nil, nil, err
	}
	// Wait 3 seconds for an ARP reply.
	for {
		if time.Since(start) > time.Second*3 {
			return nil, nil, errors.New("timeout getting ARP reply")
		}
		data, _, err := handle.ReadPacketData()
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

func send(l ...gopacket.SerializableLayer) error {
	if err := gopacket.SerializeLayers(buffer, opts, l...); err != nil {
		return err
	}

	return handle.WritePacketData(buffer.Bytes())
}
