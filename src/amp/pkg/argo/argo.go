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
	Bool
	Byte
	ByteSlice
	Complex64
	Complex128
	Dict
	Float32
	Float64
	Int32
	Int64
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

var (
	bigintMagicNumber1, _ = big.NewIntString("8258175")
	bigintMagicNumber2, _ = big.NewIntString("8323072")
	bigint1, _            = big.NewIntString("1")
	bigint253, _          = big.NewIntString("253")
	bigint254, _          = big.NewIntString("254")
	bigint255, _          = big.NewIntString("255")
)

var typeNames = map[byte]string{
	Nil:         "nil",
	Any:         "interface{}",
	BigDecimal:  "*big.Decimal",
	Bool:        "bool",
	Byte:        "byte",
	ByteSlice:   "[]byte",
	Complex64:   "complex64",
	Complex128:  "complex128",
	Dict:        "map[string]",
	Float32:     "float32",
	Float64:     "float64",
	Int32:       "int32",
	Int64:       "int64",
	Map:         "map",
	Slice:       "[]",
	String:      "string",
	StringSlice: "[]string",
	Struct:      "struct",
	StructInfo:  "*struct.Info",
	Uint32:      "uint32",
	Uint64:      "uint64",
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
