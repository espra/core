// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package url

import (
	"net/url"
)

func Quote(s string) string {
	return urlEscape(s, false)
}

func QuoteCustom(s string, safe []byte) string {
	return urlEscapeCustom(s, false, safe)
}

func QuotePlus(s string) string {
	return urlEscape(s, true)
}

func QuotePlusCustom(s string, safe []byte) string {
	return urlEscapeCustom(s, true, safe)
}

func Unquote(s string) (string, error) {
	return urlUnescape(s, false)
}

func UnquotePlus(s string) (string, error) {
	return urlUnescape(s, true)
}

// Everything except: A-Z a-z 0-9 _.-/~!()*'
func shouldEscape(c byte) bool {
	if c <= ' ' || c >= 0x7F {
		return true
	}
	switch c {
	case '<', '>', '#', '%', '"',
		'{', '}', '|', '\\', '^', '[', ']', '`',
		'?', '&', '=', '+',
		'@', ';', ':', '$', ',':
		return true
	}
	return false
}

// When safe is nil, everything except: A-Z a-z 0-9 _.-
func shouldEscapeCustom(c byte, safe []byte) bool {
	for _, elem := range safe {
		if c == elem {
			return false
		}
	}
	if c <= ',' || c >= 0x7B {
		return true
	}
	switch c {
	case '<', '>', '#', '%', '"',
		'\\', '^', '[', ']', '`',
		'?', '&', '=', '+', '@',
		';', ':', '/':
		return true
	}
	return false
}

// The following is adapted from the standard library's http package. It is
// Copyright 2009 The Go Authors. All rights reserved. And is governed by a
// BSD-style license that can be found in the Go LICENSE file.

func ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

func urlUnescape(s string, doPlus bool) (string, error) {
	// Count %, check that they're well-formed.
	n := 0
	hasPlus := false
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			n++
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				s = s[i:]
				if len(s) > 3 {
					s = s[0:3]
				}
				return "", url.EscapeError(s)
			}
			i += 3
		case '+':
			hasPlus = doPlus
			i++
		default:
			i++
		}
	}

	if n == 0 && !hasPlus {
		return s, nil
	}

	t := make([]byte, len(s)-2*n)
	j := 0
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			t[j] = unhex(s[i+1])<<4 | unhex(s[i+2])
			j++
			i += 3
		case '+':
			if doPlus {
				t[j] = ' '
			} else {
				t[j] = '+'
			}
			j++
			i++
		default:
			t[j] = s[i]
			j++
			i++
		}
	}
	return string(t), nil
}

func urlEscape(s string, doPlus bool) string {
	spaceCount, hexCount := 0, 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			if c == ' ' && doPlus {
				spaceCount++
			} else {
				hexCount++
			}
		}
	}

	if spaceCount == 0 && hexCount == 0 {
		return s
	}

	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == ' ' && doPlus:
			t[j] = '+'
			j++
		case shouldEscape(c):
			t[j] = '%'
			t[j+1] = "0123456789abcdef"[c>>4]
			t[j+2] = "0123456789abcdef"[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

func urlEscapeCustom(s string, doPlus bool, safe []byte) string {
	spaceCount, hexCount := 0, 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscapeCustom(c, safe) {
			if c == ' ' && doPlus {
				spaceCount++
			} else {
				hexCount++
			}
		}
	}

	if spaceCount == 0 && hexCount == 0 {
		return s
	}

	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == ' ' && doPlus:
			t[j] = '+'
			j++
		case shouldEscapeCustom(c, safe):
			t[j] = '%'
			t[j+1] = "0123456789abcdef"[c>>4]
			t[j+2] = "0123456789abcdef"[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}
