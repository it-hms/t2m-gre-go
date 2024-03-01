// Copyright 2024 HMS, Inc. All rights reserved.
//
// Use of this source code is governed by proprietary license.

// t2mbcastgre implements GRE encapsulation for broadcast communcation through the talk2m
// service. Invoke this application with two flags:
//   - lan The LAN interface name
//   - tap The tap interface name that is connected to talk2m
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/google/gopacket/examples/util"
	"github.com/it-hms/t2m-gre-go/gre"
	"github.com/it-hms/t2m-gre-go/iface"
)

// Network Interface name.
type iFaceName string

func (i *iFaceName) String() string {
	return fmt.Sprint(*i)
}

// Set implementation validates the value parameter matches an interface name found on the system.
// Returns error should the interface name not be found.
func (i *iFaceName) Set(value string) error {
	infs, err := net.Interfaces()
	if err != nil {
		return err
	}
	names := iface.GetNames(infs)
	if !slices.Contains(names, value) {
		return fmt.Errorf("network interface %v not found. The host interfaces are: %v", value, names)
	}
	*i = iFaceName(value)
	return nil
}

// BPF filter for TAP interface.
const TapBPFFilter string = "dst 255.255.255.255 and not arp"

// BPF filter for LAN interace.
const LanBPFFilter string = "(dst 255.255.255.255 and not arp ) or  (ether[12:2] == 0x8892 and ether[14:2]==0xfeff)"

func main() {
	var lan iFaceName
	var tap iFaceName
	flag.Var(&lan, "lan", "interface name for LAN network, e.g. eth0")
	flag.Var(&tap, "tap", "interface name for LAN network, e.g. tap1")
	serverActivated := flag.Bool("server", false, "Server inferface activated")
	flag.Parse()
	if !*serverActivated && (lan == "" || tap == "") {
		panic("Missing flag(s); Application must be invoked with both -lan and -tap flags. Or with -server flag.")
	}
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	defer util.Run()

	runTap := make(chan bool)
	runLan := make(chan bool)

	quitTap := make(chan bool)
	quitLan := make(chan bool)

	if *serverActivated {
		server := &http.Server{
			Addr: ":8095",
		}
		log.Printf("Starting server on port 8095\n")
		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			if err := server.Close(); err != nil {
				log.Fatalf("Server close error: %v", err)
			}
		}()
		http.HandleFunc("/start", gre.GetStartHandler(LanBPFFilter, TapBPFFilter, runLan, runTap, quitLan, quitTap))
		http.HandleFunc("/stop", gre.GetStopHandler(runTap, runLan, quitLan, quitTap))
		if err := server.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}

	} else {

		lanHandle, err := iface.IfaceSetup(string(lan), LanBPFFilter)
		if err != nil {
			log.Fatalf("Fatal error occured when setting up interface %s %v \n", string(lan), err.Error())
		}
		lanIface := iface.GetHardwareIfaceByName(string(lan))
		defer lanHandle.Close()

		tapHandle, err := iface.IfaceSetup(string(tap), TapBPFFilter)
		if err != nil {
			log.Fatalf("Fatal error occured when setting up interface %s %v \n", string(lan), err.Error())
		}
		defer tapHandle.Close()

		tapHandler := gre.GetTapPacketHandle(lanHandle, lanIface)
		lanHandler := gre.GetLanPacketHandle(tapHandle)

		go func() {
			<-sigs
			log.Println("Application received SIGINT or SIGTERM; will close")
			runTap <- false
			runLan <- false
			os.Exit(0)
		}()

		go gre.PktSourceHandle(tapHandle, runTap, quitTap, tapHandler, tap.String())
		go gre.PktSourceHandle(lanHandle, runLan, quitLan, lanHandler, lan.String())
		<-quitLan
		<-quitTap
	}
}
