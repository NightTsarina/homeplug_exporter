package main

import "net"

const (
	etherType = 0x88e1

	hpVersion = 0

	nwInfoReq = 0xa038
	nwInfoCnf = 0xa039
)

var (
	hpVendor = [...]byte{0x00, 0xb0, 0x52}
)

// macAddr is a convenience type that is strictly limited to the Ethernet MAC address length,
// unlike the standard library net.HardwareAddr type.
type macAddr [6]byte

func (m macAddr) String() string {
	return net.HardwareAddr(m[:]).String()
}

type networkID [7]byte

func (n networkID) String() string {
	return net.HardwareAddr(n[:]).String()
}
