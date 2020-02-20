// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Package crypto defines common elements for cryptographic constructs.
package crypto

import (
	"errors"
	"io"
)

// Error values.
var (
	ErrReadAtUnsupported = errors.New("crypto: this XOF does not support ReadAt")
)

// DummyXOF implements the XOF interface and will return errors on all method
// calls.
var DummyXOF XOF = dummyXOF{}

type dummyXOF struct{}

func (d dummyXOF) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (d dummyXOF) ReadAt(p []byte, offset uint64) (int, error) {
	return 0, ErrReadAtUnsupported
}
