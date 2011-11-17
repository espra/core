// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	"bufio"
	"io"
	// "math"
	"reflect"
)

type Decoder struct {
	alloc int
	pad32 []byte
	pad64 []byte
	r     io.Reader
	limit int
	stack []int
}

func (dec *Decoder) Decode(v interface{}) (err error) {

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return Error("you can only decode into pointer values")
	}

	dec.alloc = 0
	dec.decode(rv)
	return

}

func (dec *Decoder) decode(rv reflect.Value) {
}

func (dec *Decoder) IgnoreNext() (err error) {
	return
}

func (dec *Decoder) Limit(v int) *Decoder {
	dec.limit = v
	return dec
}

// 	data := make([]byte, 1)
// 	_, err = dec.r.Read(data)
// 	if err != nil {
// 		return
// 	}

// 	elem := rv.Elem()
// 	got := data[0]

// 	switch elem.Kind() {
// 	case reflect.Complex64:
// 		if got != Complex64 {
// 			raise(typeError("complex64", got))
// 		}
// 	case reflect.Float32:
// 		if got == Float32 {
// 			elem.SetFloat(float64(dec.ReadFloat32()))
// 			return
// 		}
// 		raise(typeError("float32", got))
// 	case reflect.Float64:
// 		switch got {
// 		case Float64:
// 			elem.SetFloat(dec.ReadFloat64())
// 		case Float32:
// 			elem.SetFloat(float64(dec.ReadFloat32()))
// 		default:
// 			raise(typeError("float64", got))
// 		}
// 		return
// 	}

// 	return

// }

// func (dec *Decoder) ReadFloat32() float32 {
// 	_, err = dec.r.Read(dec.pad32)
// 	if err != nil {
// 		error(err)
// 	}
// 	return math.Float32frombits(uint32(data[3]) | uint32(data[2])<<8 | uint32(data[1])<<16 | uint32(data[0])<<24)
// }

// func (dec *Decoder) ReadFloat64() float64 {
// 	_, err = dec.r.Read(dec.pad64)
// 	if err != nil {
// 		error(err)
// 	}
// 	return math.Float64frombits(uint64(data[7]) | uint64(data[6])<<8 | uint64(data[5])<<16 | uint64(data[4])<<24 | uint64(data[3])<<32 | uint64(data[2])<<40 | uint64(data[1])<<48 | uint64(data[0])<<56), nil
// }

// func (dec *Decoder) ReadInt64() (value int64, err error) {
// 	val, err := dec.ReadUint64()
// 	if err != nil {
// 		return
// 	}
// 	if val&1 != 0 {
// 		return ^int64(val >> 1), nil
// 	}
// 	return int64(val >> 1), nil
// }

// func (dec *Decoder) ReadSize() (value int, err error) {
// 	bitShift := uint(0)
// 	lowByte := 1
// 	data := make([]byte, 1)
// 	for lowByte > 0 {
// 		n, err := dec.r.Read(data)
// 		if n != 1 {
// 			if err == nil {
// 				return value, Error("Couldn't read from the data stream.")
// 			}
// 			return 0, err
// 		}
// 		byteValue := int(data[0])
// 		lowByte = byteValue & 128
// 		value += (byteValue & 127) << bitShift
// 		bitShift += 7
// 	}
// 	return value, nil
// }

// func (dec *Decoder) ReadString() (value string, err error) {
// 	stringSize, err := dec.ReadSize()
// 	if err != nil {
// 		return
// 	}
// 	item := make([]byte, stringSize)
// 	n, err := dec.r.Read(item)
// 	if n != stringSize {
// 		if err == nil {
// 			return value, Error("Couldn't read from the data stream.")
// 		}
// 		return
// 	}
// 	return string(item), err
// }

// func (dec *Decoder) ReadStringSlice() (value []string, err error) {
// 	arraySize, err := dec.ReadSize()
// 	if err != nil {
// 		return
// 	}
// 	value = make([]string, arraySize)
// 	var i int
// 	for i < arraySize {
// 		stringSize, err := dec.ReadSize()
// 		if err != nil {
// 			return nil, err
// 		}
// 		item := make([]byte, stringSize)
// 		n, err := dec.r.Read(item)
// 		if n != stringSize {
// 			if err == nil {
// 				return value, Error("Couldn't read from the data stream.")
// 			}
// 			return nil, err
// 		}
// 		value[i] = string(item)
// 		i++
// 	}
// 	return value, nil
// }

// func (dec *Decoder) ReadUint64() (value uint64) {
// 	bitShift := uint64(0)
// 	lowByte := 1
// 	data := make([]byte, 1)
// 	for lowByte > 0 {
// 		n, err := dec.r.Read(data)
// 		if n != 1 {
// 			if err == nil {
// 				return value, Error("Couldn't read from the data stream.")
// 			}
// 			return 0, err
// 		}
// 		byteValue := int(data[0])
// 		lowByte = byteValue & 128
// 		value += uint64(byteValue&127) << bitShift
// 		bitShift += 7
// 	}
// 	return value, nil
// }

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		limit: 10485760, // 10 megs
		pad32: make([]byte, 4),
		pad64: make([]byte, 8),
		r:     bufio.NewReader(r),
	}
}
