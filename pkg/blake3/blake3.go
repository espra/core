// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Package blake3 implements the BLAKE3 hash function.
//
// See https://github.com/BLAKE3-team/BLAKE3 for more info.
package blake3

// This code is adapted from the BLAKE3 public domain reference implementation.

import (
	"errors"

	"dappui.com/pkg/crypto"
)

const (
	blockSize = 64
	chunkSize = 1024
	maxHeight = 54
	keySize   = 32
)

const (
	flagChunkStart uint32 = 1 << iota
	flagChunkEnd
	flagParent
	flagRoot
	flagKeyedHash
	flagDeriveKeyContext
	flagDeriveKeyMaterial
)

var (
	errInvalidKeySize = errors.New("blake3: supplied key is not 256 bits long")
)

var _ crypto.Hash = (*hash)(nil)

var iv = [8]uint32{
	0x6a09e667,
	0xbb67ae85,
	0x3c6ef372,
	0xa54ff53a,
	0x510e527f,
	0x9b05688c,
	0x1f83d9ab,
	0x5be0cd19,
}

type hash struct {
	counter int
	key     [8]uint32
	flags   uint32
	written int
}

func (h *hash) BlockSize() int {
	return blockSize
}

func (h *hash) Clone() crypto.Hash {
	ret := *h
	return &ret
}

func (h *hash) Reset() {
	h.counter = 0
	h.written = 0
}

func (h *hash) Size() int {
	return 32
}

func (h *hash) Sum(b []byte) []byte {
	return nil
}

func (h *hash) Write(p []byte) (int, error) {
	h.write(p)
	return len(p), nil
}

func (h *hash) XOF() crypto.XOF {
	return nil
}

func (h *hash) write(p []byte) {
	for len(p) > 0 {
		if h.written == chunkSize {
			h.counter++
			h.written = 0
		}
		left := chunkSize - h.written
		if left > len(p) {
			left = len(p)
		}
		p = p[left:]
		h.written += left
	}
}

// DeriveKey instantiates an instance of BLAKE3 that's suitable for key
// derivation. The supplied context identifier should be hardcoded, globally
// unique, application-specific, and not contain any variable data. Custom key
// material can be provided by writing to the returned hash, before reading the
// derived key from the XOF.
func DeriveKey(context string) crypto.Hash {
	h := &hash{
		key:   iv,
		flags: flagDeriveKeyContext,
	}
	h.write([]byte(context))
	bytes32ToWords(h.Sum(nil), &h.key)
	h.Reset()
	h.flags = flagDeriveKeyMaterial
	return h
}

// Keyed instantiates a keyed instance of BLAKE3 for use in constructing MACs
// and custom PRFs. The supplied key must be exactly 256 bits long, i.e. 32
// bytes.
func Keyed(key []byte) (crypto.Hash, error) {
	if len(key) != keySize {
		return nil, errInvalidKeySize
	}
	h := &hash{
		flags: flagKeyedHash,
	}
	bytes32ToWords(key, &h.key)
	return h, nil
}

// New instantiates an instance of BLAKE3.
func New() crypto.Hash {
	return &hash{
		key: iv,
	}
}

func bytes32ToWords(b []byte, w *[8]uint32) {
	_ = b[31] // bounds check hint to the compiler
	w[0] = uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	w[1] = uint32(b[4]) | uint32(b[5])<<8 | uint32(b[6])<<16 | uint32(b[7])<<24
	w[2] = uint32(b[8]) | uint32(b[9])<<8 | uint32(b[10])<<16 | uint32(b[11])<<24
	w[3] = uint32(b[12]) | uint32(b[13])<<8 | uint32(b[14])<<16 | uint32(b[15])<<24
	w[4] = uint32(b[16]) | uint32(b[17])<<8 | uint32(b[18])<<16 | uint32(b[19])<<24
	w[5] = uint32(b[20]) | uint32(b[21])<<8 | uint32(b[22])<<16 | uint32(b[23])<<24
	w[6] = uint32(b[24]) | uint32(b[25])<<8 | uint32(b[26])<<16 | uint32(b[27])<<24
	w[7] = uint32(b[28]) | uint32(b[29])<<8 | uint32(b[30])<<16 | uint32(b[31])<<24
}
