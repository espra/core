// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package encoding

func EncodeBase32(src []byte) (result []byte) {
	return
}

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
