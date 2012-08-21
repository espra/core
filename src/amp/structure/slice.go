// Public Domain (-) 2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package structure

func InStringSlice(xs []string, s string) bool {
	for _, e := range xs {
		if e == s {
			return true
		}
	}
	return false
}
