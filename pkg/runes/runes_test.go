// Public Domain (-) 2021-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

package runes

import (
	"bytes"
	"testing"
)

var (
	saveLen int
)

var (
	textLarge = []rune("万難定団介問点退外落止泰上年広連。般真泉事危択際接徴外都載夜再。険芋健行掲進陥表月数的災著索暮著終見者事。即阜必得迅同面望勢題転四。治終雪都質佐更認幅関億運訃況彼。献続決芸県載木対造高細言青。索時権解情科元動立視月役地関行夕。席案市善済故刑目井疑発分質外町書天高曝図。覧取権徹明懸訴局利間両保録代降芸女全。")
	textSmall = []rune("Hello, 世界")
)

type testrunes struct {
	runes []rune
}

func BenchmarkToBytes(b *testing.B) {
	l := 0
	for i := 0; i < b.N; i++ {
		l += len(ToBytes(textLarge))
		l += len(ToBytes(textSmall))
	}
	saveLen = l
}

func BenchmarkToBytesSimple(b *testing.B) {
	l := 0
	for i := 0; i < b.N; i++ {
		l += len([]byte(string(textLarge)))
		l += len([]byte(string(textSmall)))
	}
	saveLen = l
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
