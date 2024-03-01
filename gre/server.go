// Copyright 2024 HMS, Inc. All rights reserved.
//
// Use of this source code is governed by proprietary license.

// handle provides utility functions to configure interface BPF handles.
package gre

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/it-hms/t2m-gre-go/iface"
)

var handlerStarted bool = false

func checkIfaces(lanIfaceName string, wanIfaceName string) error {
	var err error = nil

	infs, err := net.Interfaces()
	if err != nil {
		return err
	}
	names := iface.GetNames(infs)
	if !slices.Contains(names, lanIfaceName) {
		err = fmt.Errorf("interface %v not found", lanIfaceName)
	}
	if !slices.Contains(names, wanIfaceName) {
		if err != nil {
			return fmt.Errorf("interfaces %v and %v not found", lanIfaceName, wanIfaceName)
		}
		err = fmt.Errorf("interface %v not found", wanIfaceName)
	}
	return err
}

func GetStopHandler(runTap chan bool, runLan chan bool, quitLan chan bool, quitTap chan bool) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received /stop request\n")

		if handlerStarted {
			runTap <- false
			runLan <- false

			time.Sleep(2 * time.Millisecond)
			select {
			case <-quitTap:
			case <-quitLan:
			default:
			}

			log.Println("Stop handler coroutines.")
			io.WriteString(w, "handler coroutines stopped!\n")
		} else {
			log.Println("Handler coroutines already stopped.")
			io.WriteString(w, "handler coroutines already stopped!\n")
		}
		handlerStarted = false
	}

}

func GetStartHandler(LanBPFFilter string, TapBPFFilter string, runLan chan bool, runTap chan bool, quitLan chan bool, quitTap chan bool) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received /start request\n")
		bdy, err := io.ReadAll(r.Body)
		log.Println("server request")
		log.Println(string(bdy[:]))
		if err != nil {
			fmt.Println(err)
		}
		var dat map[string]interface{}
		if err := json.Unmarshal(bdy, &dat); err != nil {
			log.Println(string(bdy[:]))
			http.Error(w, "HTTP body parse error, expected JSON "+err.Error(), 400)
			return
		}

		lan, ok_lan := dat["lan"].(string)
		tap, ok_tap := dat["tap"].(string)

		if !ok_tap || !ok_lan {
			log.Printf("Unexpected start request %s\n", bdy)
			http.Error(w, `Unexpected request format, expected JSON body with lan and wan keys: {"lan":"eth", "tap":"tap0"}`, 400)
			return
		}
		if lan == "" || tap == "" {
			log.Printf("Unexpected start request %s\n", bdy)
			http.Error(w, fmt.Sprintf("Unexpected request, did not find %v %v", lan, tap), 400)
			return
		}

		err = checkIfaces(lan, tap)
		if err != nil {
			log.Printf("Start request error: %s\n", err.Error())
			http.Error(w, fmt.Sprintf("Unable to start GRE server due to interface not found: %v", err.Error()), 400)
			return
		}

		lanHandle, err := iface.IfaceSetup(string(lan), LanBPFFilter)
		if err != nil {
			log.Printf("A fatal error occured durn setup of %s,  %s", string(lan), err)
			http.Error(w, fmt.Sprintf("Unable to start GRE server due to pcap handle error: %v", err.Error()), 400)
			return
		}

		tapHandle, err := iface.IfaceSetup(string(tap), TapBPFFilter)
		if err != nil {
			log.Printf("A fatal error occured durn setup of %s,  %s", string(tap), err)
			http.Error(w, fmt.Sprintf("Unable to start GRE server due to pcap handle error: %v", err.Error()), 400)
			return
		}

		if handlerStarted {
			runTap <- false
			runLan <- false
			time.Sleep(2 * time.Millisecond)
			select {
			case <-quitTap:
			case <-quitLan:
			default:
			}
		}

		lanIface := iface.GetHardwareIfaceByName(string(lan))

		tapHandler := GetTapPacketHandle(lanHandle, lanIface)
		lanHandler := GetLanPacketHandle(tapHandle)

		go PktSourceHandle(tapHandle, runTap, quitTap, tapHandler, tap)
		go PktSourceHandle(lanHandle, runLan, quitLan, lanHandler, lan)
		handlerStarted = true

		log.Printf("Start request success lan:%s tap:%s \n", lan, tap)
	}
}
