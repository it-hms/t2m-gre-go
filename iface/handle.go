// Copyright 2024 HMS, Inc. All rights reserved.
//
// Use of this source code is governed by proprietary license.

// handle provides utility functions to configure interface BPF handles.
package iface

import (
	"fmt"
	"time"

	"github.com/google/gopacket/pcap"
)

// Setup the pcap handle with defaults for broadcast GRE task
func IfaceSetup(iface string, filter string) (*pcap.Handle, error) {
	var handle *pcap.Handle
	var err error
	var timestampSource pcap.TimestampSource
	// Maximum snap length.
	const MAX_SNAPLEN = 65536

	inactive, err := pcap.NewInactiveHandle(iface)
	if err != nil {
		return nil, fmt.Errorf("could not create interface %v error: %v", iface, err)
	}
	defer inactive.CleanUp()
	if err = inactive.SetSnapLen(MAX_SNAPLEN); err != nil {
		return nil, fmt.Errorf("could not set snap length: %v", err)
	} else if err = inactive.SetPromisc(false); err != nil {
		return nil, fmt.Errorf("could not promiscuous mode, error: %v", err)
	} else if err = inactive.SetTimeout(time.Second); err != nil {
		return nil, fmt.Errorf("could not set timeout: %v", err)
	}
	if err := inactive.SetTimestampSource(timestampSource); err != nil {
		return nil, fmt.Errorf("could not set pcap timestamp source: %v", err)
	}
	if handle, err = inactive.Activate(); err != nil {
		return nil, fmt.Errorf("pcap activate error: %v", err)
	}
	if err = handle.SetBPFFilter(filter); err != nil {
		return nil, fmt.Errorf("pcap set BPF filter error: %v", err)
	}
	return handle, err
}
