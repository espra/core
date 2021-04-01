// Public Domain (-) 2010-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

package style

import (
	"testing"
)

var tests = []testcase{
	{Bold, "\x1b[1m"},
	{Bold | Red, "\x1b[1;31m"},
	{Bold | Red | WhiteBG, "\x1b[1;31;47m"},
	{Bold | Bright | Red, "\x1b[1;91m"},
	{Bold | Bright | Red | WhiteBG, "\x1b[1;91;107m"},
	{Bold | Bright | Red | WhiteBG | Reset, "\x1b[0m"},
	{Bold | Blink, "\x1b[1;5m"},
	{Bold | Dim, "\x1b[1m"},
	{Bold | Italic | Undercurl, "\x1b[1;3;4:3m"},
	{Bold | Italic | Undercurl | Underline, "\x1b[1;3;4:3m"},
	{Bold | Italic | Underline, "\x1b[1;3;4m"},
	{Bold | Foreground256(100), "\x1b[1;38:5:100m"},
	{Bold | Foreground256(100) | Background256(100), "\x1b[1;38:5:100;48:5:100m"},
	{Bold | Foreground256(100) | Undercurl256(100), "\x1b[1;38:5:100;4:3;58:5:100m"},
	{Bold | Background256(100), "\x1b[1;48:5:100m"},
	{Bold | Undercurl256(100), "\x1b[1;4:3;58:5:100m"},
	{Bold | ForegroundRGB(100, 90, 80), "\x1b[1;38:2::100:90:80m"},
	{Bold | BackgroundRGB(100, 90, 80), "\x1b[1;48:2::100:90:80m"},
	{Bold | UndercurlRGB(100, 90, 80), "\x1b[1;4:3;58:2::100:90:80m"},
	{Bright, ""},
	{Bright | Red, "\x1b[91m"},
	{Dim, "\x1b[2m"},
	{Invert | Italic | Strikethrough, "\x1b[3;7;9m"},
	{Reset, "\x1b[0m"},
	{Undercurl, "\x1b[4:3m"},
	{Underline, "\x1b[4m"},
}

type testcase struct {
	code Code
	want string
}

func TestCodes(t *testing.T) {
	for idx, tt := range tests {
		got := tt.code.EscapeCodes()
		if got != tt.want {
			t.Errorf("test at idx %d = %q: want %q", idx, got, tt.want)
		}
	}
}

func TestEnabled(t *testing.T) {
	got := isEnabled("0")
	if got != false {
		t.Errorf(`isEnabled("0") = %v: want false`, got)
	}
	got = isEnabled("1")
	if got != true {
		t.Errorf(`isEnabled("1") = %v: want true`, got)
	}
	got = isEnabled("2")
	if got != true {
		t.Errorf(`isEnabled("2") = %v: want true`, got)
	}
	got = Enabled()
	if got != enabled {
		t.Errorf(`Enabled() = %v: want %v`, got, enabled)
	}
}

func TestWrap(t *testing.T) {
	ori := enabled
	enabled = true
	got := Red.String()
	want := "\x1b[31m"
	if got != want {
		t.Errorf(`COLOR=1 Red.String() = %q: want %q`, got, want)
	}
	got = Wrap("test", Bold|Red)
	want = "\x1b[1;31mtest\x1b[0m"
	if got != want {
		t.Errorf(`COLOR=1 Wrap("test", Bold|Red) = %q: want %q`, got, want)
	}
	enabled = false
	got = Red.String()
	if got != "" {
		t.Errorf(`COLOR=0 Red.String() = %q: want ""`, got)
	}
	got = Wrap("test", Bold|Red)
	if got != "test" {
		t.Errorf(`COLOR=0 Wrap("test", Bold|Red) = %q: want "test"`, got)
	}
	enabled = ori
}
