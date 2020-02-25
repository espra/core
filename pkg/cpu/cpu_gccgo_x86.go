// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386,gccgo amd64,gccgo

package cpu

//extern gccgoGetCpuidCount
func gccgoGetCpuidCount(eaxArg, ecxArg uint32, eax, ebx, ecx, edx *uint32)

//extern gccgoXgetbv
func gccgoXgetbv(eax, edx *uint32)

func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32) {
	var a, b, c, d uint32
	gccgoGetCpuidCount(eaxArg, ecxArg, &a, &b, &c, &d)
	return a, b, c, d
}

func xgetbv() (eax, edx uint32) {
	var a, d uint32
	gccgoXgetbv(&a, &d)
	return a, d
}
