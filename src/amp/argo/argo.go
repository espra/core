// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package argo

import (
	"amp/big"
	"bytes"
	"os"
	"strings"
)

const (
	magicNumber int64 = 8258175
)

var (
	zero = []byte{'\x01', '\x80', '\x01', '\x01'}
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

func pow(x, y int64) (z int64) {
	var i int64
	z = 1
	for i = 0; i < y; i++ {
		z = z * x
	}
	return z
}

func WriteInt(value int64, buffer *bytes.Buffer) os.Error {
	if value == 0 {
		buffer.Write(zero)
		return nil
	}
	if value > 0 {
		if value < magicNumber {
			encoding := []byte{'\x01', '\x80', '\x01', '\x01'}
			div, mod := value/255, value%255
			encoding[3] = byte(mod) + 1
			if div > 0 {
				div, mod = div/255, div%255
				encoding[2] = byte(mod) + 1
				if div > 0 {
					encoding[1] = byte(div) + 128
				}
			}
			buffer.Write(encoding)
		} else {
			value -= magicNumber
			buffer.WriteByte('\x01')
			buffer.WriteByte('\xff')
			lead, left := value/255, value%255
			var n int64 = 1
			for (lead / pow(253, n)) > 0 {
				n += 1
			}
			buffer.WriteByte(byte(n) + 1)
			buffer.WriteByte('\xff')
			leadChars := make([]byte, 0)
			for {
				var mod int64
				if lead == 0 {
					break
				}
				lead, mod = lead/253, lead%253
				leadChars = bytes.AddByte(leadChars, byte(mod)+2)
			}
			lenLead := len(leadChars)
			if lenLead > 0 {
				for i := lenLead - 1; i >= 0; i-- {
					buffer.WriteByte(leadChars[i])
				}
			}
			if left > 0 {
				buffer.WriteByte('\x01')
				buffer.WriteByte(byte(left))
			}
		}
	} else {
		value = -value
		if value < magicNumber {
			encoding := []byte{'\x01', '\x7f', '\xfe', '\xfe'}
			div, mod := value/255, value%255
			encoding[3] = 254 - byte(mod)
			if div > 0 {
				div, mod = div/255, div%255
				encoding[2] = 254 - byte(mod)
				if div > 0 {
					encoding[1] = 127 - byte(div)
				}
			}
			buffer.Write(encoding)
		} else {
			value -= magicNumber
			buffer.WriteByte('\x01')
			buffer.WriteByte('\x00')
			lead, left := value/254, value%254
			var n int64 = 1
			for (lead / pow(253, n)) > 0 {
				n += 1
			}
			buffer.WriteByte(254 - byte(n))
			buffer.WriteByte('\x00')
			leadChars := make([]byte, 0)
			for {
				var mod int64
				if lead == 0 {
					break
				}
				lead, mod = lead/253, lead%253
				leadChars = bytes.AddByte(leadChars, 253-byte(mod))
			}
			lenLead := len(leadChars)
			if lenLead > 0 {
				for i := lenLead - 1; i >= 0; i-- {
					buffer.WriteByte(leadChars[i])
				}
			}
			if lenLead > 1 {
				buffer.WriteByte('\x00')
			}
			buffer.WriteByte('\xfe')
			if left > 0 {
				buffer.WriteByte(254 - byte(left))
			} else {
				buffer.WriteByte('\xfe')
			}
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
