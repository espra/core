// Public Domain (-) 2010-2011 The Ampify Authors.
// See the UNLICENSE file for details.

// The slice package provides utility functions for dealing with slices.
package slice

// The ``AppendString`` function takes a reference to a slice object and appends
// a new string to it -- making sure to expand the slice if it's at capacity.
func AppendString(slice *[]string, s string) {
	length := len(*slice)
	if cap(*slice) == length {
		temp := make([]string, length, 2*(length+1))
		for idx, item := range *slice {
			temp[idx] = item
		}
		*slice = temp
	}
	*slice = (*slice)[0 : length+1]
	(*slice)[length] = s
}

func AppendByte(slice *[]byte, b byte) {
	length := len(*slice)
	if cap(*slice) == length {
		temp := make([]byte, length, 2*(length+1))
		for idx, item := range *slice {
			temp[idx] = item
		}
		*slice = temp
	}
	*slice = (*slice)[0 : length+1]
	(*slice)[length] = b
}
