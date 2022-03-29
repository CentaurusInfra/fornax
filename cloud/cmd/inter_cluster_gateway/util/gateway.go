package util

import (
	"encoding/json"
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
	"k8s.io/klog/v2"

	subnetv1 "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/subnet/v1"
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

var (
	router           routing.Router
	handle           *pcap.Handle
	opts             gopacket.SerializeOptions
	buffer           gopacket.SerializeBuffer
	localHostIP      net.IP
	localGatewayHost string
)

// initialize the commandline options
func InitFlag() {
	var logLevel string
	flag.StringVar(&logLevel, "log_level", "3", "log level")

	flag.Parse()

	local := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	klog.InitFlags(local)
	err := local.Set("v", logLevel)
	if err != nil {
		fmt.Printf("error setting klog flags: %v", err)
		os.Exit(1)
	}
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
	localGatewayHost = fmt.Sprintf("%s:%d", "127.0.0.1", FirepowerPort)
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

func syncSubnet(subnet *subnetv1.Subnet, remoteGatewayIP net.IP, remoteHostIP net.IP, remoteHostMac net.HardwareAddr, remotePort, localPort int) error {
	iface, _, src, err := router.Route(remoteHostIP)
	if err != nil {
		return err
	}

	eth := layers.Ethernet{
		SrcMAC:       iface.HardwareAddr,
		DstMAC:       remoteHostMac,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip4 := layers.IPv4{
		SrcIP:    src,
		DstIP:    remoteGatewayIP,
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
	}
	tcp := layers.TCP{
		SrcPort: layers.TCPPort(localPort),
		DstPort: layers.TCPPort(remotePort),
		SYN:     true,
	}
	err = tcp.SetNetworkLayerForChecksum(&ip4)
	if err != nil {
		return err
	}
	udp := layers.UDP{
		SrcPort: layers.UDPPort(localPort),
		DstPort: layers.UDPPort(remotePort),
	}
	err = udp.SetNetworkLayerForChecksum(&ip4)
	if err != nil {
		return err
	}

	subnetByteArray, err := json.Marshal(subnet)
	if err != nil {
		return err
	}
	payload := gopacket.Payload(subnetByteArray)

	if err := send(&eth, &ip4, &tcp, &udp, &payload); err != nil {
		return err
	}

	return nil
}
