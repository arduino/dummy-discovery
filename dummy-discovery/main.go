//
// This file is part of dummy-discovery.
//
// Copyright 2021 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.
//

package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/arduino/go-properties-orderedmap"
	discovery "github.com/arduino/pluggable-discovery-protocol-handler"
	"github.com/arduino/pluggable-discovery-protocol-handler/dummy-discovery/args"
)

type DummyDiscovery struct {
	startSyncCount int
	closeChan      chan<- bool
}

func main() {
	args.ParseArgs()
	dummyDiscovery := &DummyDiscovery{}
	server := discovery.NewDiscoveryServer(dummyDiscovery)
	if err := server.Run(os.Stdin, os.Stdout); err != nil {
		os.Exit(1)
	}
}

func (d *DummyDiscovery) Hello(userAgent string, protocol int) error {
	return nil
}

func (d *DummyDiscovery) Quit() {}

func (d *DummyDiscovery) Stop() error {
	if d.closeChan != nil {
		d.closeChan <- true
		d.closeChan = nil
	}
	return nil
}

func (d *DummyDiscovery) StartSync(eventCB discovery.EventCallback, errorCB discovery.ErrorCallback) error {
	d.startSyncCount++
	if d.startSyncCount%5 == 0 {
		return errors.New("could not start_sync every 5 times")
	}

	c := make(chan bool)
	d.closeChan = c

	// Run synchronous event emitter
	go func() {
		var closeChan <-chan bool = c

		// Output initial port state
		eventCB("add", CreateDummyPort())
		eventCB("add", CreateDummyPort())

		// Start sending events
		count := 0
		for count < 2 {
			count++

			select {
			case <-closeChan:
				return
			case <-time.After(2 * time.Second):
			}

			port := CreateDummyPort()
			eventCB("add", port)

			select {
			case <-closeChan:
				return
			case <-time.After(2 * time.Second):
			}

			eventCB("remove", &discovery.Port{
				Address:  port.Address,
				Protocol: port.Protocol,
			})
		}

		errorCB("unrecoverable error, cannot send more events")
		<-closeChan
	}()

	return nil
}

var dummyCounter = 0

func CreateDummyPort() *discovery.Port {
	dummyCounter++
	return &discovery.Port{
		Address:       fmt.Sprintf("%d", dummyCounter),
		AddressLabel:  "Dummy upload port",
		Protocol:      "dummy",
		ProtocolLabel: "Dummy protocol",
		Properties: properties.NewFromHashmap(map[string]string{
			"vid": "0x2341",
			"pid": "0x0041",
			"mac": fmt.Sprintf("%d", dummyCounter*384782),
		}),
	}
}
