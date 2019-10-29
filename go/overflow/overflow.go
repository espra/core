// Public Domain (-) 2018-present, The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// Package overflow adds support for detecting overflow on integer operations.
package overflow

// MulU64 returns the result of multiplying the given uint64 values and whether
// it's okay or not in terms of overflowing.
func MulU64(a, b uint64) (v uint64, ok bool) {
	res := a * b
	if a != 0 && res/a != b {
		return res, false
	}
	return res, true
}
