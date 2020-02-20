// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kangaroo12

import (
	"dappui.com/pkg/crypto"
)

type sponge struct {
	a       [25]uint64 // main state
	buf     []byte     // points into storage
	storage storageBuf
}

func (s *sponge) Read(p []byte) (int, error) {
	s.squeeze(p)
	return len(p), nil
}

func (s *sponge) ReadAt(p []byte, offset uint64) (int, error) {
	return 0, crypto.ErrReadAtUnsupported
}

func (s *sponge) absorb(p []byte) {
	for len(p) > 0 {
		if len(s.buf) == 0 && len(p) >= rate {
			// The fast path; absorb a full "rate" bytes of input and apply the
			// permutation.
			xorIn(s, p[:rate])
			p = p[rate:]
			keccakP1600(&s.a)
		} else {
			// The slow path; buffer the input until we can fill the sponge, and
			// then XOR it in.
			left := rate - len(s.buf)
			if left > len(p) {
				left = len(p)
			}
			s.buf = append(s.buf, p[:left]...)
			p = p[left:]
			// If the sponge is full, apply the permutation.
			if len(s.buf) == rate {
				xorIn(s, s.buf)
				s.buf = s.storage.asBytes()[:0]
				keccakP1600(&s.a)
			}
		}
	}
}

func (s *sponge) clone() *sponge {
	ret := *s
	ret.buf = ret.storage.asBytes()[:len(ret.buf)]
	return &ret
}

// finalize appends the domain separation bits in dsbyte, applies the
// multi-bitrate 10..1 padding rule, and applies the permutation. The method
// expects the first "1" bit from the padding to have been merged into the
// provided dsbyte value.
func (s *sponge) finalize(dsbyte byte) {
	// Pad with the domain-separator bits. We know that there's at least one
	// byte of space in s.buf because, if it were full, the permutation would
	// have already been applied to empty it.
	s.buf = append(s.buf, dsbyte)
	zerosStart := len(s.buf)
	s.buf = s.storage.asBytes()[:rate]
	for i := zerosStart; i < rate; i++ {
		s.buf[i] = 0
	}
	// This adds the final one bit for the padding. Because of the way that
	// bits are numbered from the LSB upwards, the final bit is the MSB of
	// the last byte.
	s.buf[rate-1] ^= 0x80
	xorIn(s, s.buf)
	keccakP1600(&s.a)
	copyOut(s, s.buf)
}

func (s *sponge) init() {
	s.buf = s.storage.asBytes()[:0]
}

func (s *sponge) reset() {
	for i := range s.a {
		s.a[i] = 0
	}
	s.buf = s.storage.asBytes()[:0]
}

func (s *sponge) squeeze(p []byte) {
	for len(p) > 0 {
		n := copy(p, s.buf)
		s.buf = s.buf[n:]
		p = p[n:]
		// Apply the permutation if we've squeezed the sponge dry.
		if len(s.buf) == 0 {
			keccakP1600(&s.a)
			s.buf = s.storage.asBytes()[:rate]
			copyOut(s, s.buf)
		}
	}
}
