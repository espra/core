// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package url

import (
	"testing"
)

func TestQuote(t *testing.T) {

	s := "~tav/some bull$hit!"
	s1 := "~tav/some%20bull%24hit!"
	s2 := "~tav/some+bull%24hit!"

	if Quote(s) != s1 {
		t.Errorf("Unexpected quote for %q: %q\n", s, Quote(s))
		return
	}

	if QuotePlus(s) != s2 {
		t.Errorf("Unexpected quote plus for %q: %q\n", s, QuotePlus(s))
		return
	}

	if QuoteCustom(s, []byte{'/', '!'}) != "%7etav/some%20bull%24hit!" {
		t.Errorf("Unexpected quote custom for %q: %q\n", s,
			QuoteCustom(s, []byte{'/', '!'}))
		return
	}

	q1, _ := Unquote(s1)
	if q1 != s {
		t.Errorf("Unexpected unquote for %q: %q\n", s1, q1)
		return
	}

	q2, _ := Unquote(s2)
	if q2 != "~tav/some+bull$hit!" {
		t.Errorf("Unexpected unquote for %q: %q\n", s2, q2)
		return
	}

	q3, _ := UnquotePlus(s2)
	if q3 != s {
		t.Errorf("Unexpected unquote plus for %q: %q\n", s2, q3)
		return
	}

}
