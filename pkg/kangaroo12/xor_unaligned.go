// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 amd64 ppc64le

package kangaroo12

import (
	"unsafe"
)

type storageBuf [rate / 8]uint64

func (s *storageBuf) asBytes() *[rate]byte {
	return (*[rate]byte)(unsafe.Pointer(s))
}

func copyOut(s *sponge, buf []byte) {
	ab := (*[rate]byte)(unsafe.Pointer(&s.a[0]))
	copy(buf, ab[:])
}

func xorIn(s *sponge, buf []byte) {
	n := len(buf)
	bw := (*[rate / 8]uint64)(unsafe.Pointer(&buf[0]))[: n/8 : n/8]
	if n >= 72 {
		s.a[0] ^= bw[0]
		s.a[1] ^= bw[1]
		s.a[2] ^= bw[2]
		s.a[3] ^= bw[3]
		s.a[4] ^= bw[4]
		s.a[5] ^= bw[5]
		s.a[6] ^= bw[6]
		s.a[7] ^= bw[7]
		s.a[8] ^= bw[8]
	}
	if n >= 104 {
		s.a[9] ^= bw[9]
		s.a[10] ^= bw[10]
		s.a[11] ^= bw[11]
		s.a[12] ^= bw[12]
	}
	if n >= 136 {
		s.a[13] ^= bw[13]
		s.a[14] ^= bw[14]
		s.a[15] ^= bw[15]
		s.a[16] ^= bw[16]
	}
	if n >= 144 {
		s.a[17] ^= bw[17]
	}
	if n >= 168 {
		s.a[18] ^= bw[18]
		s.a[19] ^= bw[19]
		s.a[20] ^= bw[20]
	}
}
