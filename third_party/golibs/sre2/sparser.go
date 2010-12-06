package sre2

// Describes a string parser type. Notably, allows users to examine the current
// rune, peek at the next rune, consume literals between prefix/suffix, and
// rebase the cursor in an absolute fashion within the underlying string. On
// instantiation, this string parser is focused 'before' the initial string:
// calling curr() will return -1, and peek() will return the first rune.
//
// Within sre2, this is used for both the parser as well as matchers to
// traverse through the input string. The curr()/peek() semantics are most
// useful for identifying conditions between runes, such as '\W', '\w' or '$'
// and '^' in multiline mode.

import (
	"strings"
	"utf8"
)

type SafeReader struct {
	str  string // backing string
	ch   int    // current ch
	opos int    // previous (absolute) position in str, before ch
	pos  int    // current (absolute) position in str, after ch
}

func NewSafeReader(str string) SafeReader {
	return SafeReader{str, -1, -1, 0}
}

// Absolute position after the current character, inside SafeReader. This will
// be -1 if EOF.
func (r *SafeReader) npos() int {
	return r.pos
}

// Returns the current focus rune in SafeReader. This will be -1 if EOF or if
// nextCh() has not yet been called.
func (r *SafeReader) curr() int {
	return r.ch
}

// Peek at the next focus rune in SafeReader.
func (r *SafeReader) peek() int {
	if r.pos < len(r.str) {
		rune, _ := utf8.DecodeRuneInString(r.str[r.pos:])
		return rune
	}
	return -1
}

// Move forward, and return the next rune. This will return -1 if the string is
// at EOF.
func (r *SafeReader) nextCh() int {
	if r.pos < len(r.str) {
		rune, size := utf8.DecodeRuneInString(r.str[r.pos:])
		r.ch = rune
		r.opos = r.pos
		r.pos += size
	} else {
		r.ch = -1
		r.opos = r.pos
		r.pos = -1
	}
	return r.ch
}

// Refocus the parser at a given point within the parsed string. When this method
// returns, the current focus rune will be directly after the given index.
func (r *SafeReader) jump(to int) {
	r.pos = to
	r.nextCh()
}

// Consume a known literal at the given point. If the literal does not exist,
// starting with the current focus rune, then panic.
func (r *SafeReader) consume(str string) {
	if r.opos == -1 {
		panic("can't consume before reading")
	}
	if !strings.HasPrefix(r.str[r.opos:], str) {
		panic("could not find correct str: " + str + " in available: " + r.str[r.opos:])
	}
	r.jump(r.opos + len(str))
}

// Consume a literal that *must* exist, between the given prefix and suffix.
// Searches for the prefix starting with the current focus rune, and returns
// this SafeReader focused on the rune directly after the suffix.
// e.g. "abcde\Qhello\Eblah", with literal("\\Q", "\\E"), returns "hello" 
//    focus --^                         and will focus on "b" after "\E".
func (r *SafeReader) literal(prefix string, suffix string) string {
	r.consume(prefix)
	start := r.opos

	idx := strings.Index(r.str[r.opos:], suffix)
	if idx == -1 {
		panic("could not find correct suffix: " + suffix)
	}
	r.jump(r.opos + idx + len(suffix))

	return r.str[start : start+idx]
}
