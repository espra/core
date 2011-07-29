// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package structure

import (
	"testing"
)

func TestSortedKeys(t *testing.T) {
	dict := map[string]string{
		"tav": "espian",
		"james": "arthur",
		"sofia": "bustamante",
		"mamading": "ceesay",
	}
	keys := SortedKeys(dict)
	expected := []string{"james", "mamading", "sofia", "tav"}
	for idx, key := range keys {
		if key != expected[idx] {
			t.Errorf("Unexpected item at position %d: %s (expected: %s)",
				idx, key, expected[idx])
		}
	}

}
