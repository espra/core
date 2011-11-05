// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	"amp/big"
	"os"
)

const (
	Nil = iota
	Any
	BigDecimal
	BigInt
	Bool
	BoolFalse
	BoolTrue
	Byte
	ByteSlice
	Complex64
	Complex128
	Dict
	Float32
	Float64
	Header
	Int32
	Int64
	Item
	Map
	Slice
	String
	StringSlice
	Struct
	StructInfo
	Uint32
	Uint64
	sentinel
)

const magicNumber int64 = 8258175

var (
	bigintMagicNumber1, _ = big.NewIntString("8258175")
	bigintMagicNumber2, _ = big.NewIntString("8323072")
	bigint1, _            = big.NewIntString("1")
	bigint253, _          = big.NewIntString("253")
	bigint254, _          = big.NewIntString("254")
	bigint255, _          = big.NewIntString("255")
	zero                  = []byte{'\x01', '\x80', '\x01', '\x01'}
	zeroBase              = []byte{'\x80', '\x01', '\x01'}
)

type Error string

func (err Error) String() string {
	return "argo error: " + string(err)
}

var OutOfRangeError = Error("out of range size value")

type TypeMismatchError string

func (err TypeMismatchError) String() string {
	return "argo error: " + string(err)
}

var typeNames = map[byte]string{
	Nil:         "nil",
	Any:         "interface{}",
	BigDecimal:  "big.Decimal",
	BigInt:      "big.Int",
	Bool:        "bool",
	BoolFalse:   "bool",
	BoolTrue:    "bool",
	Byte:        "byte",
	ByteSlice:   "[]byte",
	Complex64:   "complex64",
	Complex128:  "complex128",
	Dict:        "map[string]",
	Float32:     "float32",
	Float64:     "float64",
	Header:      "rpc.Header",
	Int32:       "int32",
	Int64:       "int64",
	Item:        "Item",
	Map:         "map",
	Slice:       "[]interface{}",
	String:      "string",
	StringSlice: "[]string",
	Struct:      "struct",
	StructInfo:  "structInfo",
	Uint32:      "uint32",
	Uint64:      "uint64",
}

func typeError(expected string, got byte) os.Error {
	return TypeMismatchError("expected " + expected + ", got " + typeNames[got])
}

func init() {
	if sentinel > baseId {
		panic("argo: type IDs have been allocated beyond the base limit")
	}
}
