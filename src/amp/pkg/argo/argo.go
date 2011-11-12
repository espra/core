// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	"amp/big"
)

const (
	Nil = iota
	Any
	BigDecimal
	BigDecimalSlice
	BigInt
	BigIntSlice
	Bool
	BoolSlice
	BoolFalse
	BoolTrue
	Byte
	ByteSlice
	ByteSliceSlice
	Complex64
	Complex64Slice
	Complex128
	Complex128Slice
	Dict
	DictAny
	DictAnySlice
	Float32
	Float32Slice
	Float64
	Float64Slice
	Header
	Int32
	Int32Slice
	Int64
	Int64Slice
	Item
	ItemSlice
	Map
	Slice
	SliceAny
	String
	StringSlice
	Struct
	StructSlice
	StructInfo
	Uint32
	Uint32Slice
	Uint64
	Uint64Slice
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

var typeNames = map[byte]string{
	Nil:            "nil",
	Any:            "interface{}",
	BigDecimal:     "big.Decimal",
	BigInt:         "big.Int",
	Bool:           "bool",
	BoolFalse:      "bool",
	BoolTrue:       "bool",
	Byte:           "byte",
	ByteSlice:      "[]byte",
	ByteSliceSlice: "[][]byte",
	Complex64:      "complex64",
	Complex128:     "complex128",
	Dict:           "map[string]",
	Float32:        "float32",
	Float64:        "float64",
	Header:         "rpc.Header",
	Int32:          "int32",
	Int64:          "int64",
	Item:           "Item",
	Map:            "map",
	Slice:          "[]interface{}",
	String:         "string",
	StringSlice:    "[]string",
	Struct:         "struct",
	StructInfo:     "structInfo",
	Uint32:         "uint32",
	Uint64:         "uint64",
}

func init() {
	if sentinel > baseId {
		panic("argo: type IDs have been allocated beyond the base limit")
	}
}
