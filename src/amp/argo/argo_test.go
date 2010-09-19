// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package argo

import (
	"bytes"
	"testing"
)

func Buffer() *bytes.Buffer {
	return bytes.NewBuffer([]byte{})
}

func TestWriteSize(t *testing.T) {

	tests := map[uint64]string{
		0:                    "\x00",
		123456789:            "\x95\x9a\xef:",
		18446744073709551615: "\xff\xff\xff\xff\xff\xff\xff\xff\xff\x01",
	}

	for value, expected := range tests {
		buf := Buffer()
		WriteSize(value, buf)
		if string(buf.Bytes()) != expected {
			t.Errorf("Got unexpected encoding for %d: %q", value, buf.Bytes())
		}
	}

}

func BenchmarkWriteSize(b *testing.B) {
	buf := Buffer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		WriteSize(123456789, buf)
	}
}

func BenchmarkWriteNumber(b *testing.B) {
	buf := Buffer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		WriteNumber("123456789", buf)
	}
}
