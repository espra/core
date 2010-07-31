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
	hashTableFactor = 16
	hashTableSize   = 1 << hashTableFactor
	maxLiteralRef   = 1 << 5
	maxOffset       = 1 << 13
	maxRef          = (1 << 8) + (1 << 3)
)


func Compress(input []byte) (output []byte, outputSize int, err os.Error) {

	inputSize := len(input)
	if inputSize <= 4 {
		return nil, 0, err
	}

	output = make([]byte, inputSize)
	hashTable := make([]byte, hashTableSize)

	var hval, idx, ref uint32

	literal := 0
	ref = 0
	inputPos := 0
	outputPos := 1
	hval = (uint32(input[0]) << 8) | uint32(input[1])

	for inputPos < (inputSize - 2) {

		hval = (hval << 8) | uint32(input[2])
		idx = ((hval >> (24 - hashTableFactor)) - (hval * 5)) & (hashTableSize - 1)
		ref = hashTable[idx]
		hashTable[idx] = inputPos
		offset := inputPos - backref - 1

		if (ref < inputPos) && (offset < maxOffset) && (inputPos + 4 < inputSize) && (ref > 0) && (ref == input[inputPos]) {

			length := 2
			maxLength := inputSize - inputPos - length

			if maxLength > maxRef {
				maxLength = maxRef
			} else {
				maxLength = maxLength
			}

			if (outputPos + 4) >= inputSize {
				if (outputPos - !literal + 4) >= inputSize {
					return nil, 0, err
				}
			}

			output[-literal - 1] = literal - 1
			outputPos  -= !literal

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
