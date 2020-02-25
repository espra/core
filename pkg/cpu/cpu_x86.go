// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

// +build 386 amd64

package cpu

import (
	"encoding/binary"
)

func has(v uint32, bitpos uint32) bool {
	return v&(1<<bitpos) != 0
}

func init() {
	var (
		stateXMM bool
		stateYMM bool
		stateZMM bool
	)

	max, ebx, ecx, edx := cpuid(0, 0)
	manufacturer := make([]byte, 12)
	binary.LittleEndian.PutUint32(manufacturer, ebx)
	binary.LittleEndian.PutUint32(manufacturer[4:], edx)
	binary.LittleEndian.PutUint32(manufacturer[8:], ecx)

	switch string(manufacturer) {
	case "AuthenticAMD":
		Manufacturer = "AMD"
	case "GenuineIntel":
		Manufacturer = "Intel"
	}

	if max >= 1 {
		_, _, ecx, _ := cpuid(1, 0)
		OSXSAVE = has(ecx, 27)
		SSE41 = has(ecx, 19)
		if OSXSAVE {
			// If OSXSAVE=1, then XSAVE=1 implicitly, so we don't need to also
			// check for that.
			eax, _ := xgetbv()
			// Check that the OS supports the various register states.
			stateXMM = has(eax, 1)
			stateYMM = has(eax, 2) && stateXMM
			stateZMM = has(eax, 3) && stateYMM
			if !stateXMM {
				// Since an OS can enable SSE support without setting OSXSAVE
				// (via FXSAVE), we test for it independently, and only disable
				// it if XMM support has not been enabled.
				SSE41 = false
			}
			if stateYMM {
				AVX = has(ecx, 28)
			}
		}
	}

	if max >= 7 {
		_, ebx, _, _ := cpuid(7, 0)
		if stateYMM {
			AVX2 = has(ebx, 5)
		}
		if stateZMM {
			AVX512F = has(ebx, 16)
		}
	}
}
