// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Package kangaroo12 implements the KangarooTwelve hash function.
//
// See https://keccak.team/kangarootwelve.html for more info.
package kangaroo12

import (
	"encoding/binary"

	"dappui.com/pkg/crypto"
)

const (
	chunkSize = 8192
	rate      = (1600 - 256) / 8
)

var _ crypto.Hash = (*k12)(nil)

type k12 struct {
	buf     [32]byte
	chunk   sponge
	count   int
	custom  []byte
	meta    sponge
	written int
}

func (k *k12) BlockSize() int {
	return rate
}

func (k *k12) Clone() crypto.Hash {
	return k.clone()
}

func (k *k12) Reset() {
	k.chunk.reset()
	k.meta.reset()
	k.count = 0
	k.written = 0
}

func (k *k12) Size() int {
	return 32
}

func (k *k12) Sum(b []byte) []byte {
	s := k.finalize()
	buf := make([]byte, 32)
	s.squeeze(buf)
	return append(b, buf...)
}

func (k *k12) Write(p []byte) (int, error) {
	k.write(p)
	return len(p), nil
}

func (k *k12) XOF() crypto.XOF {
	return k.finalize()
}

func (k *k12) clone() *k12 {
	ret := *k
	ret.chunk.buf = ret.chunk.storage.asBytes()[:len(ret.chunk.buf)]
	ret.meta.buf = ret.meta.storage.asBytes()[:len(ret.meta.buf)]
	return &ret
}

func (k *k12) finalize() *sponge {
	// Fast-path small messages.
	if k.count == 0 && (k.written+len(k.custom)+9) < chunkSize {
		meta := k.meta.clone()
		meta.absorb(k.custom)
		meta.absorb(encodeLength(len(k.custom)))
		meta.finalize(0x07)
		return meta
	}
	k = k.clone()
	k.write(k.custom)
	k.write(encodeLength(len(k.custom)))
	if k.count == 0 {
		k.meta.finalize(0x07)
	} else {
		k.chunk.finalize(0x0b)
		k.chunk.squeeze(k.buf[:])
		k.meta.absorb(k.buf[:])
		k.meta.absorb(encodeLength(k.count))
		k.meta.absorb([]byte{0xff, 0xff})
		k.meta.finalize(0x06)
	}
	return &k.meta
}

// Adapted from David Wong's implementation of Kangaroo12:
// https://github.com/mimoo/GoKangarooTwelve
func (k *k12) write(p []byte) {
	for len(p) > 0 {
		if k.written == chunkSize {
			if k.count == 0 {
				k.meta.absorb([]byte{3, 0, 0, 0, 0, 0, 0, 0})
			} else {
				k.chunk.finalize(0x0b)
				k.chunk.squeeze(k.buf[:])
				k.meta.absorb(k.buf[:])
				k.chunk.reset()
			}
			k.count++
			k.written = 0
		}
		left := chunkSize - k.written
		if left > len(p) {
			left = len(p)
		}
		if k.count == 0 {
			k.meta.absorb(p[:left])
		} else {
			k.chunk.absorb(p[:left])
		}
		p = p[left:]
		k.written += left
	}
}

// Custom instantiates an instance of Kangaroo12 with the given identifier as
// the customization string.
func Custom(identifier string) crypto.Hash {
	h := &k12{
		custom: []byte(identifier),
	}
	h.chunk.init()
	h.meta.init()
	return h
}

// New instantiates an instance of Kangaroo12.
func New() crypto.Hash {
	h := &k12{}
	h.chunk.init()
	h.meta.init()
	return h
}

// Adapted from David Wong's implementation of Kangaroo12:
// https://github.com/mimoo/GoKangarooTwelve
func encodeLength(n int) []byte {
	var (
		buf    [9]byte
		offset int
	)
	if n == 0 {
		offset = 8
	} else {
		binary.BigEndian.PutUint64(buf[0:], uint64(n))
		for offset = 0; offset < 9; offset++ {
			if buf[offset] != 0 {
				break
			}
		}
	}
	buf[8] = byte(8 - offset)
	return buf[offset:]
}

func init() {
	crypto.RegisterHash(crypto.Kangaroo12, New)
}
