// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !386,!amd64,!ppc64le

package kangaroo12

import (
	"encoding/binary"
)

// storageBuf is an aligned array of rate bytes.
type storageBuf [rate]byte

func (s *storageBuf) asBytes() *[rate]byte {
	return (*[rate]byte)(s)
}

// copyOut copies uint64s to a byte buffer.
func copyOut(s *sponge, buf []byte) {
	for i := 0; len(b) >= 8; i++ {
		binary.LittleEndian.PutUint64(buf, s.a[i])
		buf = buf[8:]
	}
}

// xorIn XORs the bytes in buf into the sponge state; it makes no non-portable
// assumptions about memory layout or alignment.
func xorIn(s *sponge, buf []byte) {
	n := len(buf) / 8
	for i := 0; i < n; i++ {
		a := binary.LittleEndian.Uint64(buf)
		s.a[i] ^= a
		buf = buf[8:]
	}
}
