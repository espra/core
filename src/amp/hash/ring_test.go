// Public Domain (-) 2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package hash

import (
	"testing"
)

func TestRing(t *testing.T) {
	r := NewRing("foo", "bar")
	r.Remove("bar")
	r.Add("hmz")
	r.Add("foo")
	if r.Len() != 2*defaultRingWeight {
		t.Logf("Mismatching bucket length: expected %d, got %d.", 2*defaultRingWeight, r.Len())
	}
	for _, key := range []string{"hello", "world", "yes"} {
		if _, ok := r.FindString(key); !ok {
			t.Logf("Error finding match for %s", key)
		}
	}
	r.Remove("hmz")
	r.Remove("foo")
	if r.Len() != 0 {
		t.Logf("Mismatching bucket length: expected 0, got %d.", r.Len())
	}
}
