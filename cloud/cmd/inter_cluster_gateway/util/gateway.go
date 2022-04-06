package util

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
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

func Send(l ...gopacket.SerializableLayer) error {
	if err := gopacket.SerializeLayers(buffer, opts, l...); err != nil {
		return err
	}

	return handle.WritePacketData(buffer.Bytes())
}

// GetNextHopHwAddr() finds out the next-hop hardware address if we want to send the packet to the destination IP
func GetNextHopHwAddr(destIP net.IP) (net.HardwareAddr, net.IP, error) {
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
	if err := Send(&eth, &arp); err != nil {
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
