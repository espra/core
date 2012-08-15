// Public Domain (-) 2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package hash

import (
	"testing"
)

func TestRing(t *testing.T) {
	r := NewRing("server-1", "server-2")
	r.Remove("server-2")
	r.Add("server-3")
	r.Add("server-1")
	if r.buckets.Len() != 2*defaultBucketsPerNode {
		t.Logf("Mismatching bucket length: expected %d, got %d.", 2*defaultBucketsPerNode, r.buckets.Len())
	}
	for _, key := range []string{"key-1", "key-2", "key-3"} {
		if _, ok := r.Find([]byte(key)); !ok {
			t.Logf("Error finding match for %s", key)
		}
	}
	if results, _ := r.FindMultiple([]byte("key"), 2); len(results) != 2 {
		t.Logf("Mismatching FindMultiple results length: expected 2, got %d", len(results))
	}
	r.Remove("server-3")
	r.Remove("server-1")
	if r.buckets.Len() != 0 {
		t.Logf("Mismatching bucket length: expected 0, got %d.", r.buckets.Len())
	}
}
