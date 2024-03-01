// Copyright 2024 HMS, Inc. All rights reserved.
//
// Use of this source code is governed by proprietary license.

// handle provides utility functions to configure interface BPF handles.
package gre

import (
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"github.com/it-hms/t2m-gre-go/dcp"
)

// Write packet to the specified handle.
func writePacket(handle *pcap.Handle, buf []byte) error {
	if err := handle.WritePacketData(buf); err != nil {
		log.Printf("Failed to send packet: %s\n", err)
		return err
	}
	return nil
}

func packetWriter(handle *pcap.Handle, buf chan []byte) {
	for {
		packet := <-buf
		if err := handle.WritePacketData(packet); err != nil {
			log.Printf("Failed to send packet: %s\n", err)
		}
	}
}

func setupWriter(handle *pcap.Handle) chan []byte {
	c := make(chan []byte)
	go packetWriter(handle, c)
	return c
}

// Returns gopacket.DataSource that is configured to decode Ethernet packets.
func commonDecode(handle gopacket.PacketDataSource) *gopacket.PacketSource {
	const decoder = "Ethernet"
	var dec gopacket.Decoder
	var ok bool
	if dec, ok = gopacket.DecodersByLayerName[decoder]; !ok {
		log.Fatalln("No decoder named", decoder)
	}

	source := gopacket.NewPacketSource(handle, dec)
	source.Lazy = true
	source.NoCopy = true
	source.DecodeStreamsAsDatagrams = true
	return source
}

func GetLanPacketHandle(writeHandle *pcap.Handle) func(packet gopacket.Packet) {
	writeChan := setupWriter(writeHandle)
	// GRE type enumeration for L2 packets
	const GreL2Type = 0x0cb20cb2
	return func(packet gopacket.Packet) {

		if dhcpLayer := packet.Layer(layers.LayerTypeDHCPv4); dhcpLayer != nil {
		} else if arp := packet.Layer(layers.LayerTypeARP); arp != nil {
		} else {
			res := EncapsulatePacket(packet.Data(), GreL2Type)
			writeChan <- res.Data()
		}
	}

}

func GetTapPacketHandle(writeHandle *pcap.Handle, writeIface *net.Interface) func(packet gopacket.Packet) {
	writeChan := setupWriter(writeHandle)
	return func(packet gopacket.Packet) {
		if greLayer, ok := packet.Layer(layers.LayerTypeGRE).(*layers.GRE); ok {
			if dcp.IsDcp(greLayer) {
				if p, err := dcp.DcpMangleSrc(greLayer.Payload, writeIface); err == nil {
					writeChan <- p.Data()
				} else {
					log.Println(err)
				}
			} else {
				writeChan <- greLayer.Payload
			}
		}
	}
}

func PktSourceHandle(tapSrc gopacket.PacketDataSource, run chan bool, quit chan bool, packetHandle func(packet gopacket.Packet), name string) {
	source := commonDecode(tapSrc)
	log.Printf("%s handler coroutine started.\n", name)
	for {
		select {
		case doRun := <-run:
			if !doRun {
				log.Printf("%s handler coroutine stop cmd received.\n", name)
				tapSrc.(*pcap.Handle).Close()
				quit <- true
				return
			}
		case p := <-source.Packets():
			if p != nil {
				packetHandle(p)
			}
		}
	}
}
