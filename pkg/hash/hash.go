// Package hash defines common elements for hash functions.
package hash

import (
	"encoding"
	"io"
)

// XOF represents an eXtendable-Output Function, which is a variable-length hash
// function in which the output can be extended to any desired length.
type XOF interface {
	// Clone returns a copy of the XOF in its current state.
	Clone() XOF
	// MarshalBinary encodes the current state of the XOF into a binary
	// representation.
	encoding.BinaryMarshaler
	// Read reads more output from the hash. Reading affects the XOF state, so
	// it's illegal to write more input after reading.
	io.Reader
	// Resets the XOF to its initial state.
	Reset()
	// Write absorbs more data into the XOF state. It panics if input is written
	// to it after output has been read from it.
	io.Writer
	// UnmarshalBinary sets the state of the XOF from the given binary
	// representation.
	encoding.BinaryUnmarshaler
}
