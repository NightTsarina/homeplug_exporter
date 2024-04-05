package main

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHomeplugFrame(t *testing.T) {
	// NW_INFO.CNF response header only.
	buf := []byte{0x00, 0x39, 0xa0, 0x00, 0xb0, 0x52}

	want := HomeplugFrame{
		Version: 0,
		MMEType: nwInfoCnf,
		Vendor:  [3]byte{0x00, 0xb0, 0x52},
		Payload: []uint8{},
	}

	hf := HomeplugFrame{}
	if err := hf.UnmarshalBinary(buf); err != nil {
		t.Errorf("unmarshal error: %v", err)
	}

	if diff := cmp.Diff(want, hf); diff != "" {
		t.Errorf("unmarshal mismatch (-want +got):\n%s", diff)
	}
}

func TestHomeplugNetworkInfo(t *testing.T) {
	// NW_INFO.CNF response payload for one network with two stations.
	buf := []byte{
		0x01, 0x2f, 0x1a, 0x52, 0x87, 0x7a, 0x78, 0x05,
		0x0c, 0x01, 0x02, 0x00, 0x0b, 0x3b, 0x5f, 0x28,
		0x52, 0x01, 0x02, 0x00, 0x0b, 0x3b, 0x5f, 0x28,
		0x56, 0x03, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0x12, 0x19, 0x00, 0x0b, 0x3b, 0x5f, 0x28, 0x72,
		0x04, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0c,
		0x10,
	}

	want := HomeplugNetworkInfo{
		Networks: []HomeplugNetworkStatus{
			{
				NetworkID:  net.HardwareAddr{0x2f, 0x1a, 0x52, 0x87, 0x7a, 0x78, 0x05},
				ShortID:    0x0c,
				TEI:        1,
				Role:       2,
				CCoAddress: net.HardwareAddr{0x00, 0x0b, 0x3b, 0x5f, 0x28, 0x52},
				CCoTEI:     1,
			},
		},
		Stations: []HomeplugStationStatus{
			{
				Address:        net.HardwareAddr{0x00, 0x0b, 0x3b, 0x5f, 0x28, 0x56},
				TEI:            3,
				BridgedAddress: net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				TxRate:         18,
				RxRate:         25,
			},
			{
				Address:        net.HardwareAddr{0x00, 0x0b, 0x3b, 0x5f, 0x28, 0x72},
				TEI:            4,
				BridgedAddress: net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				TxRate:         12,
				RxRate:         16,
			},
		},
	}

	hni := HomeplugNetworkInfo{}
	if err := hni.UnmarshalBinary(buf); err != nil {
		t.Errorf("unmarshal error: %v", err)
	}

	if diff := cmp.Diff(want, hni); diff != "" {
		t.Errorf("unmarshal mismatch (-want +got):\n%s", diff)
	}
}
