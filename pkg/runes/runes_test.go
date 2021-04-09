// Public Domain (-) 2021-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

package runes

import (
	"bytes"
	"testing"
)

type testrunes struct {
	runes []rune
}

func TestToBytes(t *testing.T) {
	for _, tt := range []testrunes{
		{[]rune("Hello, 世界")},
		{[]rune{-1, 2047, 0xd800, 0x0010fff0, 0x0f10ffff}},
	} {
		got := ToBytes(tt.runes)
		want := []byte(string(tt.runes))
		if !bytes.Equal(got, want) {
			t.Errorf("string(ToBytes([]rune(%q))) = %q", string(tt.runes), string(got))
		}
	}
}
