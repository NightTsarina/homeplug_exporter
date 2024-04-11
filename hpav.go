package main

import (
	"fmt"
	"net"
	"time"

	"github.com/mdlayher/ethernet"
	"github.com/mdlayher/packet"
)

const (
	// HomePlug AV EtherType, as defined in section 11.1.4 of the HomePlug AV specification.
	etherType = 0x88e1

	// Management message maximum length, as defined in section 11.1.7 of the HomePlug AV
	// specification, which is also ETHER_MAX_LEN as per <net/ethernet.h>.
	mmMaxLen = 1518

	// HomePlug AV management message versions, as defined in section 11.1.5 of the HomePlug AV
	// specification.
	hpavVersion1_0 = 0x00 // Version 1.0
	hpavVersion1_1 = 0x01 // Version 1.1

	// Management Message Types, as defined in section 11.1.6 of the HomePlug AV specification.
	mmTypeReq = 0b00 // Request
	mmTypeCnf = 0b01 // Confirm
	mmTypeInd = 0b10 // Indication
	mmTypeRsp = 0b11 // Response

	// HPAV raw socket read timeout.
	readTimeout = 50 * time.Millisecond
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

func readFrame(conn *packet.Conn) (*ethernet.Frame, error) {
	// The HomePlug AV specification limits the size of management messages to 1518 bytes.
	b := make([]byte, mmMaxLen)

	conn.SetReadDeadline(time.Now().Add(readTimeout))
	n, _, err := conn.ReadFrom(b)
	if err != nil {
		return nil, err
	}

	var f ethernet.Frame
	if err := (&f).UnmarshalBinary(b[:n]); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Ethernet frame: %w", err)
	}

	return &f, nil
}
