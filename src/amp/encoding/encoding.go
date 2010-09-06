// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package encoding

// Cheap integer to fixed-width decimal ASCII. Use a negative width to avoid
// zero-padding.
func PadInt(i int, width int) string {
	u := uint(i)
	if u == 0 && width <= 1 {
		return "0"
	}
	// Assemble the decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; u > 0 || width > 0; u /= 10 {
		bp--
		width--
		b[bp] = byte(u%10) + '0'
	}
	return string(b[bp:])
}

func PadInt64(i int64, width int) string {
	u := uint(i)
	if u == 0 && width <= 1 {
		return "0"
	}
	// Assemble the decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; u > 0 || width > 0; u /= 10 {
		bp--
		width--
		b[bp] = byte(u%10) + '0'
	}
	return string(b[bp:])
}
