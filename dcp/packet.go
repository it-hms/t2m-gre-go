// Copyright 2024 HMS, Inc. All rights reserved.
//
// Use of this source code is governed by proprietary license.

// dcp/packet provides function to identify and mangle DCP packets
package dcp

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// Check for T2M DCP or layer 2 encapsulation.
func IsDcp(greLayer *layers.GRE) bool {
	// T2M GRE key for DCP or Layer 2 traffic
	const GREKeyDCP = 0xFFFFFFFF
	if greLayer.Key == GREKeyDCP {
		return true
	}
	return false
}

// Create a new packet and modify the source MAC address
func DcpMangleSrc(payload []byte, iface *net.Interface) (gopacket.Packet, error) {
	packet := gopacket.NewPacket(payload, layers.LayerTypeEthernet, gopacket.Default)
	data := packet.Data()
	SetRequesterMACAddr(data[6:12])
	// change the src mac
	copy(data[6:12], iface.HardwareAddr)
	return packet, nil
}
