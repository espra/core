// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	"amp/big"
	"reflect"
)

const magicNumber int64 = 8258175

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

var (
	bigintMagicNumber1, _ = big.NewIntString("8258175")
	bigintMagicNumber2, _ = big.NewIntString("8323072")
	bigint1, _            = big.NewIntString("1")
	bigint253, _          = big.NewIntString("253")
	bigint254, _          = big.NewIntString("254")
	bigint255, _          = big.NewIntString("255")
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

func unsafeAddr(v reflect.Value) uintptr {
	if v.CanAddr() {
		return v.UnsafeAddr()
	}
	x := reflect.New(v.Type()).Elem()
	x.Set(v)
	return x.UnsafeAddr()
}

func init() {
	if sentinel > baseId {
		panic("argo: type IDs have been allocated beyond the base limit")
	}
	switch reflect.TypeOf(int(0)).Bits() {
	case 32:
		encBaseOps[reflect.Int] = Int32
		encBaseOps[reflect.Uint] = Uint32
	case 64:
		encBaseOps[reflect.Int] = Int64
		encBaseOps[reflect.Uint] = Uint64
		is64bit = true
	default:
		panic("argo: unknown size of int/uint")
	}
}
