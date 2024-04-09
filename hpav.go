package main

import "net"

const (
	// HomePlug AV EtherType, as defined in section 11.1.4 of the HomePlug AV specification.
	etherType = 0x88e1

	// HomePlug AV management message versions, as defined in section 11.1.5 of the HomePlug AV
	// specification.
	hpavVersion1_0 = 0x00 // Version 1.0
	hpavVersion1_1 = 0x01 // Version 1.1
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

// oui defines a type for holding an Organizationally Unique Identifier, which is a 24-bit number
// that uniquely identifies a vendor, manufacturer, or other organization.
type oui [3]byte
