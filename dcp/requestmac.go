// Copyright 2024 HMS, Inc. All rights reserved.
//
// Use of this source code is governed by proprietary license.

// requestmac provides functions to manage the hardware address of where the DCP or layer 2 requests origionate.
package dcp

import (
	"errors"
	"net"
)

// MAC addrress of the requesting device.
var hwAddr net.HardwareAddr = nil

// Get the MAC address of the requesting host.
func GetRequesterMacAddr() (net.HardwareAddr, error) {
	if hwAddr != nil {
		return hwAddr, nil
	}
	return nil, errors.New("Address not yet established")
}

// Set the MAC address of the requesting host.
func SetRequesterMACAddr(addr net.HardwareAddr) {
	if hwAddr == nil {
		b := make([]byte, len(addr))
		copy(b, addr)
		hwAddr = b
	}
}
