// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64,!gccgo

package kangaroo12

// This function is implemented in keccak_amd64.s.

//go:noescape
func keccakP1600(a *[25]uint64)
