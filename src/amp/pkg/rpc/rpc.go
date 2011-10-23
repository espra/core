// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package rpc

import (
	"io"
)

type Stream struct {
	stream io.ReadCloser
	length int64
}

