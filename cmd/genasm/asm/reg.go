// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

package asm

import (
	"github.com/mmcloughlin/avo/reg"
)

// XMM represents the registers introduced by SSE.
var XMM = RegisterSet{
	n: 16,
	registers: []reg.VecPhysical{
		reg.X0, reg.X1, reg.X2, reg.X3,
		reg.X4, reg.X5, reg.X6, reg.X7,
		reg.X8, reg.X9, reg.X10, reg.X11,
		reg.X12, reg.X13, reg.X14, reg.X15,
	},
	size: 128,
}

// YMM represents the registers introduced by AVX.
var YMM = RegisterSet{
	n: 16,
	registers: []reg.VecPhysical{
		reg.Y0, reg.Y1, reg.Y2, reg.Y3,
		reg.Y4, reg.Y5, reg.Y6, reg.Y7,
		reg.Y8, reg.Y9, reg.Y10, reg.Y11,
		reg.Y12, reg.Y13, reg.Y14, reg.Y15,
	},
	size: 256,
}

// ZMM represents the registers introduced by AVX512.
var ZMM = RegisterSet{
	n: 32,
	registers: []reg.VecPhysical{
		reg.Z0, reg.Z1, reg.Z2, reg.Z3,
		reg.Z4, reg.Z5, reg.Z6, reg.Z7,
		reg.Z8, reg.Z9, reg.Z10, reg.Z11,
		reg.Z12, reg.Z13, reg.Z14, reg.Z15,
		reg.Z16, reg.Z17, reg.Z18, reg.Z19,
		reg.Z20, reg.Z21, reg.Z22, reg.Z23,
		reg.Z24, reg.Z25, reg.Z26, reg.Z27,
		reg.Z28, reg.Z29, reg.Z30, reg.Z31,
	},
	size: 512,
}

// RegisterSet represents a set of physical registers.
type RegisterSet struct {
	n         int
	registers []reg.VecPhysical
	size      int
}
