// vim:ts=2:sw=2:et:ai:sts=2
package main

import (
	"errors"
	"net"

	"github.com/alecthomas/kingpin/v2"
)

var (
	bcastMAC = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	localMAC = net.HardwareAddr{0x00, 0xb0, 0x52, 0x00, 0x00, 0x01}
)

type macAddressValue net.HardwareAddr

func (m *macAddressValue) Set(s string) error {
	if s == "broadcast" || s == "all" {
		*m = (macAddressValue)(bcastMAC)
	} else if s == "local" {
		*m = (macAddressValue)(localMAC)
	} else {
		v, err := net.ParseMAC(s)
		if err != nil {
			return err
		} else if len(v) != 6 {
			return errors.New("invalid address length")
		}
		*m = (macAddressValue)(v)
	}
	return nil
}

func (m *macAddressValue) Get() interface{} {
	return *m
}

func (m *macAddressValue) String() string {
	return string(*m)
}

func MacAddress(s kingpin.Settings) (target *net.HardwareAddr) {
	target = &net.HardwareAddr{}
	s.SetValue((*macAddressValue)(target))
	return
}
