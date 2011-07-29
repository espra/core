// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package structure

import (
	"sort"
)

func SortedKeys(dict map[string]string) (keys []string) {
	keys = make([]string, len(dict))
	i := 0
	for key, _ := range dict {
		keys[i] = key
		i += 1
	}
	sort.StringSlice(keys).Sort()
	return
}