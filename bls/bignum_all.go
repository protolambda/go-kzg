package bls

import (
	"encoding/binary"
)

func (fr *Fr) String() string {
	return FrStr(fr)
}

// Checks if a *little endian* uint256 is within the Fr modulus
func ValidFr(val [32]byte) bool {
	if val[31] == 0 { // common to just use bytes31
		return true
	}
	// modulus-1 == 00000000fffffffffe5bfeff02a4bd5305d8a10908d83933487d9d2953a7ed73  (little endian)
	// 73eda753299d7d483339d80809a1d80553bda402fffe5bfeffffffff00000000 (big endian)
	// 73eda753299d7d48 3339d80809a1d805 53bda402fffe5bfe ffffffff00000000
	if a := binary.LittleEndian.Uint64(val[24:32]); a > 0x73eda753299d7d48 {
		return false
	} else if a < 0x73eda753299d7d48 {
		return true
	}
	if b := binary.LittleEndian.Uint64(val[16:24]); b > 0x3339d80809a1d805 {
		return false
	} else if b < 0x3339d80809a1d805 {
		return true
	}
	if c := binary.LittleEndian.Uint64(val[8:16]); c > 0x53bda402fffe5bfe {
		return false
	} else if c < 0x53bda402fffe5bfe {
		return true
	}
	return binary.LittleEndian.Uint64(val[0:8]) <= 0xffffffff00000000
}
