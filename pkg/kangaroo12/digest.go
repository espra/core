// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

package kangaroo12

import (
	"encoding/binary"

	"dappui.com/pkg/crypto"
)

type digester struct {
	hashes []byte
	lbuf   [9]byte
	sponge sponge
}

func (d *digester) encodeLength(n int) []byte {
	offset := 0
	binary.BigEndian.PutUint64(d.lbuf[0:], uint64(n))
	for offset = 0; offset < 9; offset++ {
		if d.lbuf[offset] != 0 {
			break
		}
	}
	d.lbuf[8] = byte(8 - offset)
	return d.lbuf[offset:]
}

func (d *digester) Digest(data []byte, b []byte) []byte {
	var digest []byte
	if n := len(b) + 32; cap(b) >= n {
		digest = b[:n]
	} else {
		digest = make([]byte, n)
		copy(digest, b)
	}
	chunks := len(data) / chunkSize
	if chunks == 0 {
		d.sponge.reset()
		d.sponge.absorb(data)
		d.sponge.absorb([]byte{0})
		d.sponge.sum(0x07, digest[len(b):])
		return digest
	}
	hlen := chunks * 32
	if cap(d.hashes) < hlen {
		d.hashes = make([]byte, hlen)
	} else {
		d.hashes = d.hashes[:0]
	}
	end := len(data) + 1
	last := chunks * chunkSize
	for start, off := chunkSize, 0; start < end; start, off = start+chunkSize, off+32 {
		d.sponge.reset()
		if start == last {
			d.sponge.absorb(data[start:])
			d.sponge.absorb([]byte{0})
		} else {
			d.sponge.absorb(data[start : start+chunkSize])
		}
		d.sponge.sum(0x0b, d.hashes[off:off+32])
	}
	d.sponge.reset()
	d.sponge.absorb(data[:chunkSize])
	d.sponge.absorb([]byte{3, 0, 0, 0, 0, 0, 0, 0})
	d.sponge.absorb(d.hashes[:hlen])
	d.sponge.absorb(d.encodeLength(chunks))
	d.sponge.absorb([]byte{0xff, 0xff})
	d.sponge.sum(0x06, digest[len(b):])
	return digest
}

// NewDigester returns a crypto.Digester instance of Kangaroo12.
func NewDigester() crypto.Digester {
	return &digester{}
}
