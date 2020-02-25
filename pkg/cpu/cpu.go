// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Package cpu implements feature detection for various CPU architectures.
package cpu

// CPU info.
var (
	Manufacturer = "Unknown"
)

// CPU features.
var (
	AVX     bool
	AVX2    bool
	AVX512F bool
	OSXSAVE bool
	SSE41   bool
)
