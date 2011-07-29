// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package logging

import (
	"io"
)

type NetworkLogger struct {
	fallback *FileLogger
	stream   *io.Writer
	receiver chan *Record
}
