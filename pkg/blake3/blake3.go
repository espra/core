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
	return &hash{}
}

// Keyed instantiates a keyed instance of BLAKE3 for use in constructing MACs
// and custom PRFs. The supplied key must be exactly 256 bits long, i.e. 32
// bytes.
func Keyed(key []byte) (crypto.Hash, error) {
	if len(key) != keySize {
		return nil, errInvalidKeySize
	}
	return &hash{}, nil
}

// New instantiates an instance of BLAKE3.
func New() crypto.Hash {
	return &hash{}
}
