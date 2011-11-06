// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	"io"
	"math"
	"os"
	"reflect"
)

type Decoder struct {
	r io.Reader
}

func (dec *Decoder) Decode(v interface{}) (err os.Error) {

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return Error("decode only makes sense for pointer values")
	}

	return dec.decodeValue(rv)

}

func (dec *Decoder) decodeValue(rv reflect.Value) (err os.Error) {

	data := make([]byte, 1)
	_, err = dec.r.Read(data)
	if err != nil {
		return
	}

	elem := rv.Elem()
	got := data[0]

	switch elem.Kind() {
	case reflect.Complex64:
		if got != Complex64 {
			return typeError("complex64", got)
		}
	case reflect.Float32:
		if got == Float32 {
			return dec.setFloat32(elem)
		}
		return typeError("float32", got)
	case reflect.Float64:
		switch got {
		case Float64:
			return dec.setFloat64(elem)
		case Float32:
			return dec.setFloat32(elem)
		}
		return typeError("float64", got)
	}

	return

}

func (dec *Decoder) setFloat32(rv reflect.Value) (err os.Error) {
	val, err := dec.ReadFloat32()
	if err != nil {
		return
	}
	rv.SetFloat(float64(val))
	return
}

func (dec *Decoder) setFloat64(rv reflect.Value) (err os.Error) {
	val, err := dec.ReadFloat64()
	if err != nil {
		return
	}
	rv.SetFloat(val)
	return
}

func (dec *Decoder) ReadFloat32() (value float32, err os.Error) {
	data := make([]byte, 4)
	_, err = dec.r.Read(data)
	if err != nil {
		return
	}
	return math.Float32frombits(uint32(data[3]) | uint32(data[2])<<8 | uint32(data[1])<<16 | uint32(data[0])<<24), nil
}

func (dec *Decoder) ReadFloat64() (value float64, err os.Error) {
	data := make([]byte, 8)
	_, err = dec.r.Read(data)
	if err != nil {
		return
	}
	return math.Float64frombits(uint64(data[7]) | uint64(data[6])<<8 | uint64(data[5])<<16 | uint64(data[4])<<24 | uint64(data[3])<<32 | uint64(data[2])<<40 | uint64(data[1])<<48 | uint64(data[0])<<56), nil
}

func (dec *Decoder) ReadInt64() (value int64, err os.Error) {
	val, err := dec.ReadUint64()
	if err != nil {
		return
	}
	if val&1 != 0 {
		return ^int64(val >> 1), nil
	}
	return int64(val >> 1), nil
}

func (dec *Decoder) ReadSize() (value int, err os.Error) {
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
	stringSize, err := dec.ReadSize()
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
	arraySize, err := dec.ReadSize()
	if err != nil {
		return
	}
	value = make([]string, arraySize)
	var i int
	for i < arraySize {
		stringSize, err := dec.ReadSize()
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
		value[i] = string(item)
		i++
	}
	return value, nil
}

func (dec *Decoder) ReadUint64() (value uint64, err os.Error) {
	bitShift := uint64(0)
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
		value += uint64(byteValue&127) << bitShift
		bitShift += 7
	}
	return value, nil
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}
