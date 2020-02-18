// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Package crypto defines common elements for cryptographic constructs.
package crypto

import (
	"io"
)

// DummyReader implements the io.Reader interface and will return an EOF on all
// calls to Read.
var DummyReader = dummyReader{}

type dummyReader struct{}

func (d dummyReader) Read(p []byte) (int, error) {
	return 0, io.EOF
}
