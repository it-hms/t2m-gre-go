// Copyright 2024 HMS, Inc. All rights reserved.
//
// Use of this source code is governed by proprietary license.

// gre/packet provides a function to GRE encapsulate existing packets.
package gre

import (
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/it-hms/t2m-gre-go/iface"
)

// Build an new packet that encapsulates the existing packet with GRE header.
func EncapsulatePacket(packet []byte, key uint32) gopacket.Packet {
	var err error
	ifaceName := "tap0"

	var ifaceIpAddr net.IP
	//destIpAddr := net.IP{10, 36, 0, 80}
	destIpAddr := net.IP{255, 255, 255, 255}
	macIface := iface.GetHardwareIfaceByName(ifaceName)

	if ifaceIpAddr, err = iface.GetIpAddr(macIface); err != nil {
		log.Fatalln("Unable to get interface IP address")
	}

	eth := &layers.Ethernet{
		SrcMAC:       macIface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		EthernetType: 0x0800,
	}
	gre := &layers.GRE{
		KeyPresent: true,
		Protocol:   layers.EthernetTypeIPv4,
		Key:        0x0cb20cb2,
		Version:    7,
	}

	ip := &layers.IPv4{
		SrcIP:    ifaceIpAddr,
		DstIP:    destIpAddr,
		Protocol: 0x2F,
		Length:   0, // will be corrected
		IHL:      5,
		Version:  4,
		TTL:      0xFF,
	}

	buf := gopacket.NewSerializeBuffer()
	// FixLengths will correct all length
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	gopacket.SerializeLayers(buf, opts,
		eth, ip, gre, gopacket.Payload(packet),
	)
	if err != nil {
		panic(err)
	}
	return gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
}
