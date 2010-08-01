// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// This package implements the ``VERY_FAST`` variant of the LZF compression
// algorithm by Marc Alexander Lehmann.
package main

import (
	"fmt"
	"os"
)

const (
	// Change the ``hashTableFactor`` to anything from 13 all the way up to 23.
	// A larger factor leads to better compression ratios at the cost of speed,
	// and vice-versa.
	hashTableFactor uint32 = 16
	hashTableSize uint32   = 1 << hashTableFactor
	maxLiteral uint32      = 1 << 5
	maxOffset uint32       = 1 << 13
	maxRef uint32          = (1 << 8) + (1 << 3)
)


func Compress(input []byte) (output []byte, outputSize int, err os.Error) {

	var inputSize int32 = int32(len(input))

	if inputSize <= 4 {
		return nil, 0, err
	}

	output = make([]byte, inputSize)
	hashTable := make([]int64, hashTableSize)

	var hSlot, offset, reference int64
	var hVal, iIdx, oIdx uint32 
	var literal int32

	hVal = uint32((((input[0]) << 8) | input[1]))

	for iIdx < uint32(inputSize - 2) {

		hVal = (hVal << 8) | uint32(input[iIdx + 2])
		hSlot =	int64((hVal ^ (hVal << 5)) >> uint32(((24 - hashTableFactor)) - uint32(hSlot) * 5) & (hashTableSize - 1))
		reference = hashTable[hSlot]
		hashTable[hSlot] = int64(iIdx)
		offset = int64(iIdx) - reference - 1

		if ((reference > 0) && (uint32(offset) < maxOffset) && (int32(iIdx + 4) < inputSize) && (input[reference + 0] == input[iIdx + 0]) && (input[reference + 1] == input[iIdx + 1]) && (input[reference + 2] == input[iIdx + 2])) {
			var length uint32 = 2
			var maxLength uint32 = uint32(inputSize) - iIdx - length

			if maxLength > maxRef {
				maxLength = maxRef
			} else {
				maxLength = maxLength
			}

			if (int32(oIdx) + literal + 4) >= inputSize {
				return nil, 0, err
			}

			output[-literal - 1] = byte(literal - 1)
			// oIdx -= !uint32(literal)
		}

	}

	return
}

func Decompress(input []byte) (n int, output []byte, err os.Error) {
	return
}

func main() {
	s := []byte("hello world")
	fmt.Println(s)
	fmt.Println(len(s))
	fmt.Println(hashTableSize)
	Compress(s)
}
