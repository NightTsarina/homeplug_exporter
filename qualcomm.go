package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/mdlayher/ethernet"
	"github.com/mdlayher/packet"
)

const (
	qualcommHdrLen = 6

	// Qualcomm vendor-specific Management Message types.
	nwInfoReq = 0xa038
	nwInfoCnf = 0xa039
)

var (
	// Qualcomm Atheros OUI used for vendor-specific HPAV extensions.
	ouiQualcomm = oui{0x00, 0xb0, 0x52}
)

// QualcommHdr is analogous to the qualcomm_hdr struct defined by the open-plc-utils reference
// implementation (mme/mme.h). The Qualcomm header extends the standard HomePlug AV header by
// adding a 3-byte vendor OUI.
type QualcommHdr struct {
	Version uint8
	MMEType uint16
	Vendor  oui
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (h *QualcommHdr) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, h); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (h *QualcommHdr) UnmarshalBinary(b []byte) error {
	return binary.Read(bytes.NewReader(b), binary.LittleEndian, h)
}

// writeQualcommReq constructs a Qualcomm vendor-specific request and writes it to the network.
func writeQualcommReq(iface *net.Interface, conn *packet.Conn, dest net.HardwareAddr, mmType uint16) error {
	hdr := QualcommHdr{
		Version: hpavVersion1_0,
		MMEType: mmType,
		Vendor:  ouiQualcomm,
	}

	b, err := hdr.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal homeplug frame: %w", err)
	}

	f := ethernet.Frame{
		Destination: dest,
		Source:      iface.HardwareAddr,
		EtherType:   etherType,
		Payload:     b,
	}

	b, err = f.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal ethernet frame: %w", err)
	}

	if _, err = conn.WriteTo(b, &packet.Addr{HardwareAddr: dest}); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}
