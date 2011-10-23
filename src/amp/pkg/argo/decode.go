// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	"io"
	"os"
)

type Decoder struct {
	r io.Reader
}

func (dec *Decoder) Decode(v interface{}) (err os.Error) {
	return
}

func (dec *Decoder) ReadInt() (value int, err os.Error) {
	bitShift := uint(0)
	lowByte := 1
	data := make([]byte, 1)
	for lowByte > 0 {
		n, err := dec.r.Read(data)
		if n != 1 {
			if err == nil {
				return value, Error("Couldn't read from the data stream.")
			}
			return 0, err
		}
		byteValue := int(data[0])
		lowByte = byteValue & 128
		value += (byteValue & 127) << bitShift
		bitShift += 7
	}
	return value, nil
}

func (dec *Decoder) ReadString() (value string, err os.Error) {
	stringSize, err := dec.ReadInt()
	if err != nil {
		return
	}
	item := make([]byte, stringSize)
	n, err := dec.r.Read(item)
	if n != stringSize {
		if err == nil {
			return value, Error("Couldn't read from the data stream.")
		}
		return
	}
	return string(item), err
}

func (dec *Decoder) ReadStringSlice() (value []string, err os.Error) {
	arraySize, err := dec.ReadInt()
	if err != nil {
		return
	}
	var i int
	for i < arraySize {
		stringSize, err := dec.ReadInt()
		if err != nil {
			return nil, err
		}
		item := make([]byte, stringSize)
		n, err := dec.r.Read(item)
		if n != stringSize {
			if err == nil {
				return value, Error("Couldn't read from the data stream.")
			}
			return nil, err
		}
		value = append(value, string(item))
		i++
	}
	return value, nil
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}
