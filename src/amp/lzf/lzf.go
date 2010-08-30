// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// LZF Compression
// ===============
//
// This package implements the ``VERY_FAST`` variant of the LZF compression
// algorithm by Marc Alexander Lehmann. Although not as fast as libLZF which is
// extremely optimised thanks to pointer arithmetic, this still works a lot
// faster than many other compression algorithms. The compression ratio is also,
// surprisingly, often better for formats like JSON.
package main

import (
	"fmt"
)

const (

	// The ``hashTableFactor`` can be set to anything from 13 all the way up to
	// 23. A larger factor leads to better compression ratios at the cost of
	// speed and vice-versa.
	hashTableFactor uint32 = 16
	hashTableSize   uint32 = 1 << hashTableFactor
	maxLiteral      uint32 = 1 << 5
	maxOffset       uint32 = 1 << 13
	maxBackref      uint32 = (1 << 8) + (1 << 3)
	MaxSize         uint32 = (1 << 32) - 4

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
func Compress(input []byte) (output []byte) {

	inputLength := uint32(len(input))
	if inputLength <= 4 || inputLength >= MaxSize {
		return nil
	}

	var backref int64
	var diff, hslot, hval, iidx, length, literal, max, oidx, offset uint32

	output = make([]byte, inputLength)
	hashTable := make([]uint32, hashTableSize)
	hval = uint32(input[0] << 8) | uint32(input[1])
	sentinel := inputLength - 2

	oidx++

	for iidx < sentinel {

		hval = (hval << 8) | uint32(input[iidx+2])
		hslot = ((hval >> (24 - hashTableFactor)) - (hval * 5)) & (hashTableSize - 1)
		backref = hashTable[hslot]
		hashTable[hslot] = iidx
		offset = iidx - backref - 1

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
			if (oidx + 4) >= inputLength {
				// And, if so, a second -- the exact but rare test.
				if literal > 0 {
					diff = 4
				} else {
					diff = 3
				}
				if (oidx + diff) >= inputLength {
					return nil
				}
			}

			output[oidx - literal - 1] = byte(literal - 1)
			if literal == 0 {
				oidx--
			}

			for {
				length++
				if (length > max) || (input[backref+length] != input[iidx+length]) {
					break
				}
			}

			length -= 2
			iidx++

			if length < 7 {
				output[oidx] = byte((offset >> 8) + (length << 5))
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

			if iidx > inputLength - 2 {
				break
			}

			iidx -= 2

			hval = (uint32(input[iidx]) << 8 | uint32(input[iidx+1]) << 8) | uint32(input[iidx+2])
			hslot = ((hval >> (24 - hashTableFactor)) - (hval * 5)) & (hashTableSize - 1)
			hashTable[hslot] = iidx
			iidx++

			hval = (hval << 8) | uint32(input[iidx+2])
			hslot = ((hval >> (24 - hashTableFactor)) - (hval * 5)) & (hashTableSize - 1)
			hashTable[hslot] = iidx
			iidx++

			iidx -= length + 1

			for length > 0 {
				hval = (hval << 8) | uint32(input[iidx+2])
				hslot = ((hval >> (24 - hashTableFactor)) - (hval * 5)) & (hashTableSize - 1)
				hashTable[hslot] = iidx
				iidx++
				length -= 1
			}

		} else {

			if oidx >= inputLength {
				return nil
			}

			literal++
			iidx++
			oidx++

			if literal == maxLiteral {
				output[oidx - literal - 1] = byte(literal - 1)
				literal = 0
				oidx++
			}

		}

	}

	if (oidx + 3) > inputLength {
		return nil
	}

	for iidx <= inputLength {
		literal++
		iidx++
		oidx++
		if (literal == maxLiteral) {
			output[oidx - literal - 1] = byte(literal - 1)
			literal = 0
			oidx++
		}
	}

	output[oidx - literal - 1] = byte(literal - 1)
	if literal == 0 {
		oidx -= 1
	}

	return output[0:oidx]

}

// The ``Decompress`` function.
func Decompress(input []byte, size uint32) []byte {

	output := make([]byte, size, size)
	inputLength := uint32(len(input))

	var backref int64
	var ctrl, iidx, length, oidx uint32

	for {

		// Get the control byte.
		ctrl = uint32(input[iidx])
		iidx++

		// The control byte indicates a literal reference.
		if ctrl < (1 << 5) {

			ctrl++

			if oidx + ctrl > size {
				return nil
			}

			// Safety check.
			if iidx + ctrl > inputLength {
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

		// The control byte indicates a back reference.
		} else {

			length = ctrl >> 5
			backref = int64(oidx - ((ctrl & 31) << 8) - 1)

			// Safety check.
			if iidx >= inputLength {
				return nil
			}

			if length == 7 {
				length += uint32(input[iidx])
				iidx++
				// Safety check.
				if iidx >= inputLength {
					return nil
				}
			}

			backref -= int64(input[iidx])
			iidx++

			if oidx + length + 2 > size {
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

		if iidx >= inputLength {
			break
		}

	}

	return output

}

func iCompress(input []byte) (output []byte) {
	return
}

type lzfPreset struct {
	codeBits []uint8
	code     []uint16
}


func (preset *lzfPreset) Compress(input []byte) (output []byte) {
	return
}

func (preset *lzfPreset) Decompress(input []byte) (output []byte) {
	return
}

// func DictionaryEncoder(dict []byte) {
// 	preset := &lzfPreset{}
// }

// func DictionaryEncoder(dict []byte) {
// 	preset := &lzfPreset{}
// }

func main() {
	s := []byte("hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world ")
	out := Compress(s)
	fmt.Println(out)
	fmt.Println(len(s))
	fmt.Println(len(out))

	test := []byte{12, 104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100, 32, 104, 224, 132, 11, 1, 100, 32}

	fmt.Println(string(test))
	deOut := Decompress(test, 156)
	if deOut == nil {
		fmt.Println("DEDEEEEEE")
	}
	_ = deOut
	fmt.Println(string(deOut))
}
