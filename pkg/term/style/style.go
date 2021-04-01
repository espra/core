// Public Domain (-) 2010-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

// Package style provides support for styling terminal output.
//
// See the documentation for the `Enabled` function to understand how styled
// output support is determined.
package style

import (
	"os"
	"strconv"
	"strings"

	"web4.cc/pkg/term"
)

// Codes for various text effects.
const (
	Blink Code = 1 << iota
	Bold
	Bright
	Dim
	Invert
	Italic
	Reset
	Strikethrough
	Undercurl
	Underline
)

// Codes for the base foreground colors.
const (
	Black Code = (iota << 16) | (1 << 10)
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// Codes for the base background colors.
const (
	BlackBG Code = (iota << 40) | (1 << 12)
	RedBG
	GreenBG
	YellowBG
	BlueBG
	MagentaBG
	CyanBG
	WhiteBG
)

var enabled bool

// Code represents colors and text effects for styling terminal output. You can
// OR multiple codes together with the following exceptions:
//
// * If Reset is set, all other codes will be ignored.
//
// * The Bold and Dim text effects cannot be OR-ed together. If they are, only
// Bold will be applied.
//
// * The Undercurl and Underline text effects cannot be OR-ed together. If they
// are, only Undercurl will be applied.
//
// * The Bright text effect only works with the base foreground and background
// colors.
//
// * At most, only two colors can be OR-ed together: a foreground color and a
// background color, or a foreground color and an undercurl color. Any other
// color combination, e.g. a foreground with another foreground color, will
// result in undefined behavior.
type Code uint64

// NOTE(tav): The 64 bits of the Code value are structured as:
//
// - 10 bits for text effects
// - 1 bit to indicate an 8-bit foreground color
// - 1 bit to indicate a 24-bit foreground color
// - 1 bit to indicate an 8-bit background color
// - 1 bit to indicate a 24-bit background color
// - 1 bit to indicate an 8-bit undercurl color
// - 1 bit to indicate a 24-bit undercurl color
// - 24 bits for the foreground color
// - 24 bits for the background/undercurl color

// EscapeCodes returns the ANSI escape codes for the Code. This function doesn't
// pay any heed to whether styled output is enabled or not.
func (c Code) EscapeCodes() string {
	if c == 0 || c&Reset != 0 {
		return "\x1b[0m"
	}
	b := strings.Builder{}
	b.WriteString("\x1b[")
	bright := false
	undercurl := false
	written := false
	// Handle text effects.
	if c&1023 != 0 {
		// NOTE(tav): Bold and Dim are exclusive of each other.
		if c&Bold != 0 {
			b.WriteByte('1')
			written = true
		} else if c&Dim != 0 {
			b.WriteByte('2')
			written = true
		}
		if c&Bright != 0 {
			bright = true
		}
		if c&Italic != 0 {
			if written {
				b.WriteByte(';')
			}
			b.WriteByte('3')
			written = true
		}
		// NOTE(tav): Undercurl and Underline are exclusive of each other.
		if c&Undercurl != 0 {
			if written {
				b.WriteByte(';')
			}
			b.WriteString("4:3")
			undercurl = true
			written = true
		} else if c&Underline != 0 {
			if written {
				b.WriteByte(';')
			}
			b.WriteByte('4')
			written = true
		}
		if c&Blink != 0 {
			if written {
				b.WriteByte(';')
			}
			b.WriteByte('5')
			written = true
		}
		if c&Invert != 0 {
			if written {
				b.WriteByte(';')
			}
			b.WriteByte('7')
			written = true
		}
		if c&Strikethrough != 0 {
			if written {
				b.WriteByte(';')
			}
			b.WriteByte('9')
			written = true
		}
	}
	// Handle foreground colors.
	c >>= 10
	if c&3 != 0 {
		color := (c >> 6) & 0xffffff
		if c&1 != 0 {
			if color <= 8 {
				if written {
					b.WriteByte(';')
				}
				if bright {
					b.WriteByte('9')
					b.WriteByte('0' + uint8(color))
				} else {
					b.WriteByte('3')
					b.WriteByte('0' + uint8(color))
				}
				written = true
			} else if color <= 255 {
				if written {
					b.WriteByte(';')
				}
				b.WriteString("38:5:")
				b.WriteString(strconv.FormatUint(uint64(color), 10))
				written = true
			}
		} else {
			if written {
				b.WriteByte(';')
			}
			b.WriteString("38:2::")
			b.WriteString(strconv.FormatUint(uint64(color&0xff), 10))
			b.WriteByte(':')
			b.WriteString(strconv.FormatUint(uint64((color>>8)&0xff), 10))
			b.WriteByte(':')
			b.WriteString(strconv.FormatUint(uint64((color>>16)&0xff), 10))
			written = true
		}
	}
	// Handle background colors.
	c >>= 2
	if c&3 != 0 {
		color := c >> 28
		if c&1 != 0 {
			if color <= 8 {
				if written {
					b.WriteByte(';')
				}
				if bright {
					b.WriteByte('1')
					b.WriteByte('0')
					b.WriteByte('0' + uint8(color))
				} else {
					b.WriteByte('4')
					b.WriteByte('0' + uint8(color))
				}
				written = true
			} else if color <= 255 {
				if written {
					b.WriteByte(';')
				}
				b.WriteString("48:5:")
				b.WriteString(strconv.FormatUint(uint64(color), 10))
				written = true
			}
		} else {
			if written {
				b.WriteByte(';')
			}
			b.WriteString("48:2::")
			b.WriteString(strconv.FormatUint(uint64(color&0xff), 10))
			b.WriteByte(':')
			b.WriteString(strconv.FormatUint(uint64((color>>8)&0xff), 10))
			b.WriteByte(':')
			b.WriteString(strconv.FormatUint(uint64((color>>16)&0xff), 10))
			written = true
		}
	}
	// Handle undercurl colors.
	c >>= 2
	if c&3 != 0 {
		if written {
			b.WriteByte(';')
		}
		if !undercurl {
			b.WriteString("4:3;")
		}
		color := c >> 26
		if c&1 != 0 {
			b.WriteString("58:5:")
			b.WriteString(strconv.FormatUint(uint64(color), 10))
			written = true
		} else {
			b.WriteString("58:2::")
			b.WriteString(strconv.FormatUint(uint64(color&0xff), 10))
			b.WriteByte(':')
			b.WriteString(strconv.FormatUint(uint64((color>>8)&0xff), 10))
			b.WriteByte(':')
			b.WriteString(strconv.FormatUint(uint64((color>>16)&0xff), 10))
			written = true
		}
	}
	if written {
		b.WriteByte('m')
		return b.String()
	}
	return ""
}

// String returns the ANSI escape codes for the Code. If styled output is not
// enabled, this function will return an empty string.
func (c Code) String() string {
	if !enabled {
		return ""
	}
	return c.EscapeCodes()
}

// Background256 returns the background color code representing the given 256
// color value.
func Background256(v uint8) Code {
	return (Code(v) << 40) | (1 << 12)
}

// BackgroundRGB returns the background color code representing the given 24-bit
// color.
func BackgroundRGB(r uint8, g uint8, b uint8) Code {
	code := Code(r) | (Code(g) << 8) | (Code(b) << 16)
	return (code << 40) | (1 << 13)
}

// Enabled returns whether styled output is enabled. This can be forced using
// the `ForceEnable` and `ForceDisable` functions, but will default to using the
// following heuristic:
//
// * If the environment variable TERMSTYLE=0, then styled output is disabled.
//
// * If TERMSTYLE=2, then styled output is enabled.
//
// * Otherwise, styled output is only enabled if stdout is connected to a
// terminal.
//
// This heuristic is run at startup using the stdout value at `os.Stdout`.
func Enabled() bool {
	return enabled
}

// ForceDisable forces styled output to be disabled.
func ForceDisable() {
	enabled = false
}

// ForceEnable forces styled output to be enabled.
func ForceEnable() {
	enabled = true
}

// ForceWrap encloses the given text with escape codes to stylize it and then
// resets the output back to normal. Unlike Wrap, this function doesn't pay any
// heed to whether styled output is enabled or not.
func ForceWrap(s string, c Code) string {
	return c.EscapeCodes() + s + "\x1b[0m"
}

// Foreground256 returns the foreground color code representing the given 256
// color value.
func Foreground256(v uint8) Code {
	return (Code(v) << 16) | (1 << 10)
}

// ForegroundRGB returns the foreground color code representing the given 24-bit
// color.
func ForegroundRGB(r uint8, g uint8, b uint8) Code {
	code := Code(r) | (Code(g) << 8) | (Code(b) << 16)
	return (code << 16) | (1 << 11)
}

// Undercurl256 returns an undercurl color code representing the given 256 color
// value.
func Undercurl256(v uint8) Code {
	return (Code(v) << 40) | (1 << 14)
}

// UndercurlRGB returns an undercurl color code representing the given 24-bit
// color.
func UndercurlRGB(r uint8, g uint8, b uint8) Code {
	code := Code(r) | (Code(g) << 8) | (Code(b) << 16)
	return (code << 40) | (1 << 15)
}

// Wrap encloses the given text with escape codes to stylize it and then resets
// the output back to normal. If styled output is not enabled, this function
// will return the given text without any changes.
func Wrap(s string, c Code) string {
	if !enabled {
		return s
	}
	return ForceWrap(s, c)
}

func isEnabled(env string) bool {
	switch env {
	case "0":
		return false
	case "2":
		return true
	}
	return term.IsTTY(os.Stdout)
}

func init() {
	enabled = isEnabled(os.Getenv("TERMSTYLE"))
}
