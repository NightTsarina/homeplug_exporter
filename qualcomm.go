package main

import (
	"bytes"
	"encoding/binary"
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
