// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// The slice package provides utility functions for dealing with slices.
package slice

// The ``AppendString`` function takes a reference to a slice object and appends
// a new string to it -- making sure to expand the slice if it's at capacity.
func AppendString(slice *[]string, s string) {
	length := len(*slice)
	if cap(*slice) == length {
		temp := make([]string, length, (2 * length) + 1)
		for idx, item := range *slice {
			temp[idx] = item
		}
		*slice = temp
	}
	*slice = (*slice)[0:length+1]
	(*slice)[length] = s
}
