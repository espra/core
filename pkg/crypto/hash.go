// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crypto

import (
	"fmt"
	"io"
)

// The defined set of cryptographic hash algorithms that we support.
const (
	BLAKE3     HashAlgorithm = 1 + iota // import dappui.com/pkg/blake3
	Kangaroo12                          // import dappui.com/pkg/kangaroo12
	maxHash
)

var hashes = [maxHash]func() Hash{}

// Digester represents a cryptographic hash function optimized for producing
// digests.
type Digester interface {
	// Digest hashes all of the given data and appends the digest to b and
	// returns the resulting slice.
	Digest(data []byte, b []byte) []byte
}

// Hash represents a cryptographic hash function.
type Hash interface {
	// BlockSize returns the Hash's underlying block size.
	BlockSize() int
	// Clone returns a copy of the Hash in its current state.
	Clone() Hash
	// Reset resets the Hash to its initial state.
	Reset()
	// Size returns the number of bytes Sum will append.
	Size() int
	// Sum appends the digest of the current state to b and returns the
	// resulting slice. It does not change the underlying state.
	Sum(b []byte) []byte
	// Write absorbs more data into the Hash's state. It never returns an error.
	io.Writer
	// XOF returns an eXtendable-Output Function for the current state of the
	// Hash. The state of the XOF should be independent of changes to the Hash's
	// state, so that it's safe to keep writing more data after instantiating
	// the XOF.
	//
	// If a Hash does not support being treated as an XOF, then it should
	// document this fact, and return a DummyXOF to satisfy the interface.
	XOF() XOF
}

// HashAlgorithm identifies a cryptographic hash function that is implemented in
// another package.
type HashAlgorithm uint

// New instantiates a new Hash for the HashAlgorithm. It panics if the
// HashAlgorithm has not been defined within this package or if an
// implementation has not been registered via a call to RegisterHash.
func (h HashAlgorithm) New() Hash {
	if h > 0 && h < maxHash {
		if f := hashes[h]; f != nil {
			return f()
		}
		panic(fmt.Errorf("crypto: %s implementation has not been registered", h))
	}
	panic(fmt.Errorf("crypto: unknown HashAlgorithm (%d)", h))
}

func (h HashAlgorithm) String() string {
	switch h {
	case BLAKE3:
		return "BLAKE3"
	case Kangaroo12:
		return "Kangaroo12"
	default:
		return fmt.Sprintf("Unknown HashAlgorithm (%d)", h)
	}
}

// XOF represents an eXtendable-Output Function.
type XOF interface {
	// Read reads more output from the XOF. It returns io.EOF if the limit has
	// been reached.
	io.Reader
	// ReadAt reads len(p) bytes into p starting at the given offset. It returns
	// io.EOF if the limit has been reached or ErrReadAtUnsupported if the XOF
	// is not seekable, and thus does not support ReadAt.
	//
	// ReadAt should not affect nor be affected by the Read offset. Clients of
	// ReadAt should be able to execute parallel ReadAt calls on the same XOF.
	ReadAt(p []byte, offset uint64) (int, error)
}

// RegisterHash registers a function that returns a new instance of the given
// HashAlgorithm. This is intended to be called from the init function in
// packages that implement hash functions.
func RegisterHash(alg HashAlgorithm, f func() Hash) {
	if alg == 0 || alg >= maxHash {
		panic(fmt.Errorf("crypto: RegisterHash called on unknown HashAlgorithm (%d)", alg))
	}
	hashes[alg] = f
}
