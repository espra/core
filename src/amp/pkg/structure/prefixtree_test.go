// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package structure

import (
	"testing"
)

func TestPrefixTree(t *testing.T) {

	tree := NewPrefixTree()
	tree.Insert("frolic", 1)
	tree.Insert("fun", 2)
	tree.Insert("funny", 3)
	tree.Insert("feel", 4)
	tree.Insert("muse", 5)
	tree.Insert("music", 6)
	tree.Insert("melody", 7)
	tree.Insert("m", 8)
	tree.Insert("muse", 9)
	tree.Insert("fuse", 10)
	tree.Insert("musique", 11)

	t.Logf("%v", tree)

	val := tree.Lookup("m")
	if val != 8 {
		t.Errorf("PrefixTree lookup does not match the expected value: %v", val)
	}

	val = tree.Lookup("muse")
	if val != 9 {
		t.Errorf("PrefixTree lookup does not match the expected value: %v", val)
	}

	val = tree.Lookup("musico")
	if val != nil {
		t.Errorf("PrefixTree lookup does not match the expected value: %v", val)
	}

	if tree.Size() != 10 {
		t.Error("PrefixTree does not match the expected size")
	}

	matches := tree.MatchPrefix("musicology")
	if len(matches) != 2 {
		t.Errorf("PrefixTree match prefix returned unexpected count: %d", len(matches))
	}

	t.Logf("\nPrefix matches:\n")
	for _, match := range matches {
		t.Logf("  %#v\n", match)
	}

	tree.Delete("music")
	tree.Delete("musique")
	tree.Delete("m")
	tree.Delete("muse")
	tree.Delete("melody")
	tree.Delete("frolic")
	tree.Delete("fun")
	tree.Delete("funny")
	tree.Delete("fuse")
	tree.Delete("feel")

	t.Logf("\n%v", tree)

}
