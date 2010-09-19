// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package argo

import (
	"amp/big"
	"bytes"
	"os"
	"strings"
)

type EncodingError string

func (err EncodingError) String() string {
	return string(err)
}

func WriteSize(value uint64, buffer *bytes.Buffer) os.Error {
	for {
		leftBits := value & 127
		value >>= 7
		if value > 0 {
			leftBits += 128
		}
		buffer.WriteByte(byte(leftBits))
		if value == 0 {
			break
		}
	}
	return nil
}

func WriteNumber(value string, buffer *bytes.Buffer) os.Error {
	if strings.Count(value, ".") > 0 {
		number, ok := big.NewDecimal(value)
		if !ok {
			return EncodingError("Couldn't create a Decimal representation of " + value)
		}
		return WriteDecimal(number, buffer)
	}
	number, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return EncodingError("Couldn't create an Int representation of " + value)
	}
	return WriteBigInt(number, buffer)
}

func WriteDecimal(value *big.Decimal, buffer *bytes.Buffer) os.Error {
	buffer.WriteByte('\x01')
	return nil
}

func WriteBigInt(value *big.Int, buffer *bytes.Buffer) os.Error {
	buffer.WriteByte('\x01')
	return nil
}
