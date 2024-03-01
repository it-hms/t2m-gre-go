// Copyright 2024 HMS, Inc. All rights reserved.
//
// Use of this source code is governed by proprietary license.

// This module provides utility functions to get MAC and IP addresses from interfaces.
package iface

import (
	"errors"
	"net"
)

// Get the IP address from the net.Interface.
func GetIpAddr(iface *net.Interface) (net.IP, error) {
	addrs, err := iface.Addrs()
	var ipv4Addr net.IP
	if err == nil && len(addrs) > 0 {
		for _, addr := range addrs { // get ipv4 address
			if ipv4Addr = addr.(*net.IPNet).IP.To4(); ipv4Addr != nil {
				break
			}
		}
		if ipv4Addr != nil {
			return ipv4Addr, nil
		}
	}
	return nil, errors.New("Unable to get IP address from iface.")
}

// Get the MAC address from the interface. The interface name, e.g. 'eth0' is the paramteter passed to the function.
func GetHardwareIfaceByName(IfaceName string) *net.Interface {
	var macIface *net.Interface
	var err error
	if macIface, err = net.InterfaceByName(IfaceName); err != nil {
		panic(err)
	}
	return macIface
}

// Get the interface IP address by interface name
func GetIfaceIPAddrByName(IfaceName string) net.IP {
	var ip net.IP
	var err error
	iface := GetHardwareIfaceByName(IfaceName)
	if ip, err = GetIpAddr(iface); err != nil {
		panic(err)
	}
	return ip
}

// getNames returns a list of interface names found on the host. Similar to iproute2 cmd `ip a`
func GetNames(interfaces []net.Interface) []string {
	var names []string
	for _, v := range interfaces {
		names = append(names, v.Name)
	}
	return names
}
