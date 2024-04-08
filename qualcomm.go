package main

const (
	// Qualcomm vendor-specific Management Message types.
	nwInfoReq = 0xa038
	nwInfoCnf = 0xa039
)

var (
	// Qualcomm Atheros OUI used for vendor-specific HPAV extensions.
	ouiQualcomm = oui{0x00, 0xb0, 0x52}
)
