// Public Domain (-) 2018-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

// Package checked adds support for detecting overflows on integer operations.
package checked

// MulU64 returns the result of multiplying two uint64 values, and whether it's
// safe or overflows.
func MulU64(a, b uint64) (v uint64, ok bool) {
	res := a * b
	if a != 0 && res/a != b {
		return res, false
	}
	return res, true
}
