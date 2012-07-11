// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// LZF Compression
// ===============
//
// This package implements the ``VERY_FAST`` variant of the LZF compression
// algorithm by Marc Alexander Lehmann. Although not as fast as libLZF which is
// extremely optimised thanks to pointer arithmetic, this still works a lot
// faster than many other compression algorithms. The compression ratio is also,
// surprisingly, often better for formats like JSON.
package lzf

const (
	// The ``hashTableFactor`` can be set to anything from 13 all the way up to
	// 23. A larger hash table leads to better compression ratios at the cost of
	// speed and vice-versa.
	hashTableFactor uint32 = 16
	hashTableSize   uint32 = 1 << hashTableFactor
	maxLiteral      uint32 = 1 << 5
	maxOffset       int64  = 1 << 13
	maxBackref      uint32 = (1 << 8) + (1 << 3)

	// The maximum size of LZF encoded byte stream is limited to a relatively
	// sane size.
	MaxSize uint32 = 1 << 30
)

// The LZF compression format is composed of::
//
//       000LLLLL <L+1>               ; literal reference
//       LLLooooo oooooooo            ; back reference L
//       111ooooo LLLLLLLL oooooooo   ; back reference L+7
//
// An upcoming literal reference is identified by a control byte which has its 3
// most significant bits set to zero. The control byte also indicates the length
// of the particular reference -- which can be up to 32 bytes long.
//
// An upcoming back reference is indicated by control bytes which have their 3
// most significant bits bits set to something other than ``000``. There are two
// slightly different variants.
//
// The standard variant indicates the length of the back reference in the 3 most
// significant bits and the rest of the byte and the proceeding byte provide the
// offset value for the back reference.
//
// However, if the length is ``111``, that is 7, then it indicates the presence
// of an extended back reference. For this the proceeding byte provides the full
// value of the length and it's the next byte after that which together with the
// remainder of the first gives the offset value for the back reference.
//
func Compress(input []byte) (output []byte) {

	inputLength := uint32(len(input))
	if inputLength <= 4 || inputLength >= MaxSize {
		return nil
	}

	var offset int64
	var backref, diff, hslot, hval, iidx, length, literal, max, oidx uint32

	output = make([]byte, inputLength+4)
	hashTable := make([]uint32, hashTableSize)
	hval = uint32(input[0]<<8) | uint32(input[1])
	sentinel := inputLength - 2

	oidx = 5

	for iidx < sentinel {

		hval = (hval << 8) | uint32(input[iidx+2])
		hslot = ((hval >> (24 - hashTableFactor)) - (hval * 5)) & (hashTableSize - 1)
		backref = hashTable[hslot]
		hashTable[hslot] = iidx
		offset = int64(iidx) - int64(backref) - 1

		if (offset < maxOffset) &&
			((iidx + 4) < inputLength) &&
			(backref > 0) &&
			(input[backref] == input[iidx]) &&
			(input[backref+1] == input[iidx+1]) &&
			(input[backref+2] == input[iidx+2]) {

			length = 2
			max = inputLength - iidx - length
			if max > maxBackref {
				max = maxBackref
			}

			// First, a faster conservative test.
			if (oidx) >= inputLength {
				// And, if so, a second -- the exact but rare test.
				if literal > 0 {
					diff = 0
				} else {
					diff = 1
				}
				if (oidx - diff) >= inputLength {
					return nil
				}
			}

			output[oidx-literal-1] = byte(literal - 1)
			if literal == 0 {
				oidx--
			}

			for {
				length++
				if (length >= max) || (input[backref+length] != input[iidx+length]) {
					break
				}
			}

			length -= 2
			iidx++

			if length < 7 {
				output[oidx] = byte(uint32(offset>>8) + (length << 5))
				oidx++
			} else {
				output[oidx] = byte((offset >> 8) + (7 << 5))
				output[oidx+1] = byte(length - 7)
				oidx += 2
			}

			output[oidx] = byte(offset)
			literal = 0
			oidx += 2
			iidx += length + 1

			if iidx >= inputLength-2 {
				break
			}

			iidx -= 2

			hval = (uint32(input[iidx])<<8 | uint32(input[iidx+1])<<8) | uint32(input[iidx+2])
			hslot = ((hval >> (24 - hashTableFactor)) - (hval * 5)) & (hashTableSize - 1)
			hashTable[hslot] = iidx
			iidx++

			hval = (hval << 8) | uint32(input[iidx+2])
			hslot = ((hval >> (24 - hashTableFactor)) - (hval * 5)) & (hashTableSize - 1)
			hashTable[hslot] = iidx
			iidx++

		} else {

			if oidx >= inputLength-4 {
				return nil
			}

			output[oidx] = input[iidx]
			iidx++
			oidx++
			literal++

			if literal == maxLiteral {
				output[oidx-literal-1] = byte(literal - 1)
				literal = 0
				oidx++
			}

		}

	}

	if (oidx - 1) > inputLength {
		return nil
	}

	for iidx < inputLength {
		output[oidx] = input[iidx]
		iidx++
		oidx++
		literal++
		if literal == maxLiteral {
			output[oidx-literal-1] = byte(literal - 1)
			literal = 0
			oidx++
		}
	}

	output[oidx-literal-1] = byte(literal - 1)
	if literal == 0 {
		oidx -= 1
	}

	output[0] = byte((inputLength >> 24) & 255)
	output[1] = byte((inputLength >> 16) & 255)
	output[2] = byte((inputLength >> 8) & 255)
	output[3] = byte((inputLength >> 0) & 255)

	return output[0:oidx]

}

// The ``Decompress`` function.
func Decompress(input []byte) (output []byte) {

	inputLength := uint32(len(input))
	if inputLength <= 4 {
		return nil
	}

	var backref int64
	var ctrl, iidx, length, oidx, outputLength uint32

	outputLength = ((uint32(input[0]) << 24) | (uint32(input[1]) << 16) | (uint32(input[2]) << 8) | uint32(input[3]))

	if outputLength >= MaxSize {
		return nil
	}

	output = make([]byte, outputLength, outputLength)
	iidx = 4

	for iidx < inputLength {

		// Get the control byte.
		ctrl = uint32(input[iidx])
		iidx++

		if ctrl < (1 << 5) {

			// The control byte indicates a literal reference.
			ctrl++
			if oidx+ctrl > outputLength {
				return nil
			}

			// Safety check.
			if iidx+ctrl > inputLength {
				return nil
			}

			for {
				output[oidx] = input[iidx]
				iidx++
				oidx++
				ctrl--
				if ctrl == 0 {
					break
				}
			}

		} else {

			// The control byte indicates a back reference.
			length = ctrl >> 5
			backref = int64(oidx - ((ctrl & 31) << 8) - 1)

			// Safety check.
			if iidx >= inputLength {
				return nil
			}

			// It's an extended back reference. Read the extended length before
			// reading the full back reference location.
			if length == 7 {
				length += uint32(input[iidx])
				iidx++
				// Safety check.
				if iidx >= inputLength {
					return nil
				}
			}

			// Put together the full back reference location.
			backref -= int64(input[iidx])
			iidx++

			if oidx+length+2 > outputLength {
				return nil
			}

			if backref < 0 {
				return nil
			}

			output[oidx] = output[backref]
			oidx++
			backref++
			output[oidx] = output[backref]
			oidx++
			backref++

			for {
				output[oidx] = output[backref]
				oidx++
				backref++
				length--
				if length == 0 {
					break
				}
			}

		}

	}

	return output

}

func Preset(dict []byte) (preset *lzfPreset) {
	temp := Compress(dict)
	if temp == nil {
		return nil
	}
	preset = &lzfPreset{}
	preset.dict = dict
	preset.length = len(dict)
	preset.compressed = temp
	return
}

type lzfPreset struct {
	dict       []byte
	length     int
	compressed []byte
}

func (preset *lzfPreset) Compress(input []byte) (output []byte) {
	return
}

func (preset *lzfPreset) Decompress(input []byte) (output []byte) {
	combined := make([]byte, preset.length+len(input))
	_ = combined
	return
}
