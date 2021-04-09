// Public Domain (-) 2021-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

// Package runes implements simple functions to manipulate Unicode runes.
package runes

import (
	"unicode/utf8"
)

const (
	max  = 0x0010ffff
	max1 = 1<<7 - 1
	max2 = 1<<11 - 1
	max3 = 1<<16 - 1
	smax = 0xdfff
	smin = 0xd800
)

// ByteLen returns the byte length of the rune when encoded into UTF-8 using
// `utf8.EncodeRune`.
func ByteLen(r rune) int {
	switch {
	case r < 0:
		return 3
	case r <= max1:
		return 1
	case r <= max2:
		return 2
	case smin <= r && r <= smax:
		return 3
	case r <= max3:
		return 3
	case r <= max:
		return 4
	}
	return 3
}

// ToBytes converts a rune slice to a UTF-8 encoded byte slice.
func ToBytes(s []rune) []byte {
	l := 0
	for i := 0; i < len(s); i++ {
		l += ByteLen(s[i])
	}
	b := make([]byte, l)
	l = 0
	for i := 0; i < len(s); i++ {
		l += utf8.EncodeRune(b[l:], s[i])
	}
	return b
}
