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


func FRST(inArray []byte, inPtr uint32) uint32 {
	return uint32((((inArray[inPtr]) << 8) | inArray[inPtr + 1]))
}

func NEXT(hValue uint32, inArray []byte, inPtr uint32) uint32 {
	return uint32((((hValue) << 8) | inArray[inPtr + 2]))
}

func Compress(input []byte) (output []byte, outputSize int, err os.Error) {

	var inputSize int32 = int32(len(input))

	if inputSize <= 4 {
		return nil, 0, err
	}

	output = make([]byte, inputSize)
	hashTable := make([]int64, hashTableSize)

	var offset, reference int64
	var hSlot, hVal, iIdx, length, maxLength, oIdx uint32 
	var literal int32

	// hVal = uint32((((input[0]) << 8) | input[1]))
	// hVal = uint32((((input[iIdx]) << 8) | input[iIdx + 1]))
	hVal = FRST(input, iIdx)

	for {
		if iIdx < uint32(inputSize - 2) {

			// hVal = (hVal << 8) | uint32(input[iIdx + 2])
			hVal = NEXT(hVal, input, iIdx)
			hSlot =	(hVal ^ (hVal << 5)) >> uint32(((24 - hashTableFactor)) - uint32(hSlot) * 5) & (hashTableSize - 1)
			reference = hashTable[hSlot]
			hashTable[hSlot] = int64(iIdx)
			offset = int64(iIdx) - reference - 1

			if (reference > 0) && (uint32(offset) < maxOffset) && (int32(iIdx + 4) < inputSize) && (input[reference + 0] == input[iIdx + 0]) && (input[reference + 1] == input[iIdx + 1]) && (input[reference + 2] == input[iIdx + 2]) {
				length = 2
				maxLength = uint32(inputSize) - iIdx - length

				if maxLength > maxRef {
					maxLength = maxRef
				} else {
					maxLength = maxLength
				}

				if (oIdx + uint32(literal + 4)) >= uint32(inputSize) {
					return nil, 0, err
				}
				
				for ;; {
					length++
					if length < maxLength && input[uint32(reference) + length] == input[iIdx + uint32(literal)] {
						continue
					} else {
						break
					}
				}

				if literal != 0 {
					output[oIdx] = byte(literal - 1)
					oIdx++
					literal = -literal

					for ;; {
						output[oIdx] = input[iIdx + uint32(literal)]
						oIdx++
						literal++

						if literal != 0 {
							continue
						} else {
							break
						}
					}
				}

				length += 2
				iIdx++

				if length < 7 {
					output[oIdx] = byte(uint32(offset >> 8) + (length << 5))
					oIdx++
				} else {
					output[oIdx] = byte((offset >> 8) + (7 << 5))
					oIdx++
					output[oIdx] = byte(length - 7)
					oIdx++
				}

				output[oIdx] = byte(offset)
				oIdx++

				iIdx += length - 1
				hVal = FRST(input, iIdx)

			}

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
