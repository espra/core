// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	"amp/big"
	"amp/rpc"
	"bytes"
	"io"
	"math"
	"os"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unsafe"
	"utf8"
	"fmt"
)

const (
	baseId   = 64
	maxInt32 = 2147483647
)

const (
	rawEngine = iota
	arrayEngine
	dictEngine
	indirEngine
	mapEngine
	reflectEngine
	sliceEngine
	structEngine
)

type encEngine struct {
	elems    []*encElem
	explicit bool
	i        int /* array: length, indir: indir, raw: op, struct: length */
	rt       reflect.Type
	typ      int
	typeId   []byte
}

type encElem struct {
	engine *encEngine
	idx    int
	name   []byte
	offset uintptr
}

var encEngineRunners = map[int]func(*Encoder, *encEngine, uintptr){
	indirEngine: func(enc *Encoder, engine *encEngine, p uintptr) {
		indir := engine.i
		up := unsafe.Pointer(p)
		for indir > 0 {
			up = *(*unsafe.Pointer)(up)
			if up == nil {
				error(PointerError)
			}
			indir--
		}
		elemEngine := engine.elems[0].engine
		encEngineRunners[elemEngine.typ](enc, elemEngine, uintptr(up))
	},
	rawEngine: func(enc *Encoder, engine *encEngine, p uintptr) {
		if engine.explicit {
			enc.b.Write(engine.typeId)
		}
		encOps[engine.i](enc, p)
	},
	reflectEngine: func(enc *Encoder, engine *encEngine, p uintptr) {
		// v := (interface{})(p)
		// enc.encode(v)
		enc.encode(unsafe.Unreflect(engine.rt, unsafe.Pointer(p)))
	},
	sliceEngine: func(enc *Encoder, engine *encEngine, p uintptr) {
		if engine.explicit {
			enc.b.Write(engine.typeId)
		}
		header := (*reflect.SliceHeader)(unsafe.Pointer(p))
		elem := engine.elems[0]
		mark := header.Data
		offset := elem.offset
		size := header.Len
		elemEngine := elem.engine
		buf := enc.b
		explicit := elemEngine.explicit
		typeId := elemEngine.typeId
		if elemEngine.typ == rawEngine {
			op := encOps[elemEngine.i]
			for i := 0; i < size; i++ {
				if explicit {
					buf.Write(typeId)
				}
				op(enc, mark)
				mark += offset
			}
		} else if elemEngine.typ == indirEngine {
			// var raw bool
			subEngine := elemEngine.elems[0].engine
			runner := encEngineRunners[subEngine.typ]
			for i := 0; i < size; i++ {
				if explicit {
					buf.Write(typeId)
				}
				indir := elemEngine.i
				up := unsafe.Pointer(mark)
				for indir > 0 {
					up = *(*unsafe.Pointer)(up)
					if up == nil {
						error(PointerError)
					}
					indir--
				}
				// op(enc, mark)
				runner(enc, subEngine, uintptr(up))
				mark += offset
			}
		} else {
			runner := encEngineRunners[elemEngine.typ]
			for i := 0; i < size; i++ {
				if explicit {
					buf.Write(typeId)
				}
				runner(enc, elemEngine, mark)
				mark += offset
			}
		}
	},
	structEngine: func(enc *Encoder, engine *encEngine, p uintptr) {
		for _, elem := range engine.elems {
			field := elem.engine
			if field.explicit {
				enc.b.Write(field.typeId)
			}
			encEngineRunners[field.typ](enc, field, p+elem.offset)
		}
	},
}

var encOps = [...]func(*Encoder, uintptr){
	Int32: func(enc *Encoder, p uintptr) {
		v := *(*int32)(unsafe.Pointer(p))
		enc.writeInt64(int64(v), enc.b)
	},
	Int64: func(enc *Encoder, p uintptr) {
		v := *(*int64)(unsafe.Pointer(p))
		enc.writeInt64(v, enc.b)
	},
	String: func(enc *Encoder, p uintptr) {
		v := *(*string)(unsafe.Pointer(p))
		enc.writeSize(len(v), enc.b)
		enc.b.Write([]byte(v))
	},
	StringSlice: func(enc *Encoder, p uintptr) {
		v := *(*[]string)(unsafe.Pointer(p))
		enc.writeStringSlice(v, enc.b)
	},
}

var (
	encAny            = []byte{Any}
	encBigDecimal     = []byte{BigDecimal}
	encBigInt         = []byte{BigInt}
	encBool           = []byte{Bool}
	encBoolFalse      = []byte{Bool, 0}
	encBoolTrue       = []byte{Bool, 1}
	encByte           = []byte{Byte}
	encByteSlice      = []byte{ByteSlice}
	encByteSliceSlice = []byte{ByteSliceSlice}
	encComplex64      = []byte{Complex64}
	encComplex128     = []byte{Complex128}
	encDict           = []byte{Dict}
	encDictAny        = []byte{DictAny}
	encFloat32        = []byte{Float32}
	encFloat64        = []byte{Float64}
	encHeader         = []byte{Header}
	encInt32          = []byte{Int32}
	encInt64          = []byte{Int64}
	encItem           = []byte{Item}
	encMap            = []byte{Map}
	encNil            = []byte{Nil}
	encSlice          = []byte{Slice}
	encSliceAny       = []byte{SliceAny}
	encString         = []byte{String}
	encStringSlice    = []byte{StringSlice}
	encStruct         = []byte{Struct}
	encStructInfo     = []byte{StructInfo}
	encUint32         = []byte{Uint32}
	encUint64         = []byte{Uint64}
)

var encCompiler sync.RWMutex
var encEngines = map[reflect.Type]*encEngine{}

var encBaseOps = [255]int{
	reflect.Bool:       Bool,
	reflect.Complex64:  Complex64,
	reflect.Complex128: Complex128,
	reflect.Float32:    Float32,
	reflect.Float64:    Float64,
	reflect.Int8:       Int32,
	reflect.Int16:      Int32,
	reflect.Int32:      Int32,
	reflect.Int64:      Int64,
	reflect.String:     String,
	reflect.Uint8:      Byte,
	reflect.Uint16:     Uint32,
	reflect.Uint32:     Uint32,
	reflect.Uint64:     Uint64,
}

var encSliceOps = [255]int{
	reflect.Bool:       BoolSlice,
	reflect.Complex64:  Complex64Slice,
	reflect.Complex128: Complex128Slice,
	reflect.Float32:    Float32Slice,
	reflect.Float64:    Float64Slice,
	reflect.Int8:       Int32Slice,
	reflect.Int16:      Int32Slice,
	reflect.Int32:      Int32Slice,
	reflect.Int64:      Int64Slice,
	reflect.String:     StringSlice,
	reflect.Uint8:      ByteSlice,
	reflect.Uint16:     Uint32Slice,
	reflect.Uint32:     Uint32Slice,
	reflect.Uint64:     Uint64Slice,
}

var encSliceSliceOps = [255]int{
	reflect.Uint8: ByteSliceSlice,
}

var encTypeIndicator = [...][]byte{
	Any:            encAny,
	BigDecimal:     encBigDecimal,
	BigInt:         encBigInt,
	Bool:           encBool,
	Byte:           encByte,
	ByteSlice:      encByteSlice,
	ByteSliceSlice: encByteSliceSlice,
	Complex64:      encComplex64,
	Complex128:     encComplex128,
	Dict:           encDict,
	DictAny:        encDictAny,
	Float32:        encFloat32,
	Float64:        encFloat64,
	Header:         encHeader,
	Int32:          encInt32,
	Int64:          encInt64,
	Item:           encItem,
	Map:            encMap,
	Nil:            encNil,
	Slice:          encSlice,
	SliceAny:       encSliceAny,
	String:         encString,
	StringSlice:    encStringSlice,
	Struct:         encStruct,
	StructInfo:     encStructInfo,
	Uint32:         encUint32,
	Uint64:         encUint64,
}

var is64bit bool
var anyInterfaceType = reflect.TypeOf(interface{}(nil))

var encOps2 = [...]func(*Encoder, reflect.Value, bool){
	Bool: func(enc *Encoder, v reflect.Value, typeinfo bool) {
		if v.Bool() {
			enc.b.Write(encBoolTrue)
		} else {
			enc.b.Write(encBoolFalse)
		}
	},
	Byte: func(enc *Encoder, v reflect.Value, typeinfo bool) {
		if typeinfo {
			enc.b.Write(encByte)
		}
		enc.b.Write([]byte{byte(v.Int())})
	},
	ByteSlice: func(enc *Encoder, v reflect.Value, typeinfo bool) {
		if typeinfo {
			enc.b.Write(encByteSlice)
		}
		enc.writeByteSlice(v.Interface().([]byte), enc.b)
	},
	Complex64: func(enc *Encoder, v reflect.Value, typeinfo bool) {
		if typeinfo {
			enc.b.Write(encComplex64)
		}
		val := v.Complex()
		enc.writeFloat32(float32(real(val)), enc.b)
		enc.writeFloat32(float32(imag(val)), enc.b)
	},
	Complex128: func(enc *Encoder, v reflect.Value, typeinfo bool) {
		if typeinfo {
			enc.b.Write(encComplex128)
		}
		val := v.Complex()
		enc.writeFloat64(real(val), enc.b)
		enc.writeFloat64(imag(val), enc.b)
	},
	Float32: func(enc *Encoder, v reflect.Value, typeinfo bool) {
		if typeinfo {
			enc.b.Write(encFloat32)
		}
		enc.writeFloat32(float32(v.Float()), enc.b)
	},
	Float64: func(enc *Encoder, v reflect.Value, typeinfo bool) {
		if typeinfo {
			enc.b.Write(encFloat64)
		}
		enc.writeFloat64(v.Float(), enc.b)
	},
	Int32: func(enc *Encoder, v reflect.Value, typeinfo bool) {
		if typeinfo {
			enc.b.Write(encInt32)
		}
		enc.writeInt64(v.Int(), enc.b)
	},
	Int64: func(enc *Encoder, v reflect.Value, typeinfo bool) {
		if typeinfo {
			enc.b.Write(encInt64)
		}
		enc.writeInt64(v.Int(), enc.b)
	},
	String: func(enc *Encoder, v reflect.Value, typeinfo bool) {
		if typeinfo {
			enc.b.Write(encString)
		}
	},
	StringSlice: func(enc *Encoder, v reflect.Value, typeinfo bool) {
		if typeinfo {
			enc.b.Write(encStringSlice)
		}
		enc.writeStringSlice(v.Interface().([]string), enc.b)
	},
	Uint32: func(enc *Encoder, v reflect.Value, typeinfo bool) {
		if typeinfo {
			enc.b.Write(encUint32)
		}
		enc.writeUint64(v.Uint(), enc.b)
	},
	Uint64: func(enc *Encoder, v reflect.Value, typeinfo bool) {
		if typeinfo {
			enc.b.Write(encUint64)
		}
		enc.writeUint64(v.Uint(), enc.b)
	},
}

type Encoder struct {
	b       *bytes.Buffer
	dup     bool
	engines map[reflect.Type]*encEngine
	scratch []byte
	w       io.Writer
}

func (enc *Encoder) Encode(v interface{}) (err os.Error) {
	enc.encode(v)
	if enc.dup {
		_, err = enc.w.Write(enc.b.Bytes())
	}
	return
}

func (enc *Encoder) encode(v interface{}) {
	switch value := v.(type) {
	case *big.Decimal:
		enc.b.Write(encBigDecimal)
		val := EncodeBigDecimal(value)
		enc.writeSize(len(val), enc.b)
		enc.b.Write(val)
	case *big.Int:
		enc.b.Write(encBigInt)
		val := EncodeBigInt(value)
		enc.writeSize(len(val), enc.b)
		enc.b.Write(val)
	case bool:
		if value {
			enc.b.Write(encBoolTrue)
		} else {
			enc.b.Write(encBoolFalse)
		}
	case byte:
		enc.b.Write(encByte)
	case []byte:
		enc.b.Write(encByteSlice)
		enc.writeByteSlice(value, enc.b)
	case complex64:
		enc.b.Write(encComplex64)
		enc.writeFloat32(real(value), enc.b)
		enc.writeFloat32(imag(value), enc.b)
	case complex128:
		enc.b.Write(encComplex128)
		enc.writeFloat64(real(value), enc.b)
		enc.writeFloat64(imag(value), enc.b)
	case float32:
		enc.b.Write(encFloat32)
		enc.writeFloat32(value, enc.b)
	case float64:
		enc.b.Write(encFloat64)
		enc.writeFloat64(value, enc.b)
	case int:
		if is64bit {
			enc.b.Write(encInt64)
		} else {
			enc.b.Write(encInt32)
		}
		enc.writeInt64(int64(value), enc.b)
	case int8:
		enc.b.Write(encInt32)
		enc.writeInt64(int64(value), enc.b)
	case int16:
		enc.b.Write(encInt32)
		enc.writeInt64(int64(value), enc.b)
	case int32:
		enc.b.Write(encInt32)
		enc.writeInt64(int64(value), enc.b)
	case int64:
		enc.b.Write(encInt64)
		enc.writeInt64(value, enc.b)
	case []interface{}:
		enc.b.Write(encSliceAny)
		enc.writeSlice(value, enc.b)
	case map[string]interface{}:
		enc.b.Write(encDictAny)
		enc.writeDict(value, enc.b)
	case rpc.Header:
		enc.b.Write(encHeader)
		enc.writeDict(value, enc.b)
	case string:
		enc.b.Write(encString)
		enc.writeSize(len(value), enc.b)
		enc.b.Write([]byte(value))
	case []string:
		enc.b.Write(encStringSlice)
		enc.writeStringSlice(value, enc.b)
	case uint:
		if is64bit {
			enc.b.Write(encUint64)
		} else {
			enc.b.Write(encUint32)
		}
		enc.writeUint64(uint64(value), enc.b)
	case uint16:
		enc.b.Write(encUint32)
		enc.writeUint64(uint64(value), enc.b)
	case uint32:
		enc.b.Write(encUint32)
		enc.writeUint64(uint64(value), enc.b)
	case uint64:
		enc.b.Write(encUint64)
		enc.writeUint64(value, enc.b)
	default:
		enc.encodeValue(reflect.ValueOf(v))
	}
}

func (enc *Encoder) encodeValue(rv reflect.Value) {

	if !rv.IsValid() {
		error(Error("invalid type detected"))
	}

	rt := rv.Type()
	if enc.engines == nil {
		enc.engines = make(map[reflect.Type]*encEngine)
	} else if engine, present := enc.engines[rt]; present {
		enc.runEngine(engine, rv)
		return
	}

	encCompiler.Lock()
	engine, err := compileEncEngine(rt)
	encCompiler.Unlock()

	if err != nil {
		error(err)
	}

	enc.checkinEngine(engine)
	enc.runEngine(engine, rv)

}

func (enc *Encoder) runEngine(engine *encEngine, rv reflect.Value) {
	if !engine.explicit {
		enc.b.Write(engine.typeId)
	}
	if engine.typ == indirEngine {
		indir := engine.i
		for indir > 0 {
			rv = rv.Elem()
			indir--
		}
		elemEngine := engine.elems[0].engine
		encEngineRunners[elemEngine.typ](enc, elemEngine, unsafeAddr(rv))
	} else {
		encEngineRunners[engine.typ](enc, engine, unsafeAddr(rv))
	}
}

var _ = fmt.Printf

func (enc *Encoder) checkinEngine(engine *encEngine) {
	if engine.typ == structEngine {
		if _, present := enc.engines[engine.rt]; !present {
			enc.writeStructInfo(engine, enc.b)
			enc.engines[engine.rt] = engine
		}
	}
	if engine.elems != nil {
		for _, elem := range engine.elems {
			if elem.engine != nil && elem.engine.typ != rawEngine {
				enc.checkinEngine(elem.engine)
			}
		}
	}
}

// var forcekey bool
// buf := enc.b
// typ := v.Type()
// keytype := typ.Key()
// elemtype := typ.Elem()
// if keytype.Kind() == reflect.String {
// 	if elemtype == anyInterfaceType {
// 		buf.Write(encDictAny)
// 		enc.writeDict(v.Interface().(map[string]interface{}), buf)
// 		return
// 	} else {
// 		buf.Write(encDict)
// 	}
// } else {
// 	buf.Write(encMap)
// 	if keytype == anyInterfaceType {
// 		buf.Write(encAny)
// 		forcekey = true
// 	} else {
// 		forcekey = enc.writeEncType(keytype)
// 	}
// }
// var forceval bool
// if elemtype == anyInterfaceType {
// 	buf.Write(encAny)
// 	forceval = true
// } else {
// 	forceval = enc.writeEncType(elemtype)
// }

// for _, key := range v.MapKeys() {
// 	enc.encode(key.Interface(), forcekey)
// 	enc.encode(v.MapIndex(key).Interface(), forceval)
// }

// func (enc *Encoder) runEngine(engine *encEngine, v reflect.Value) {
// 	switch engine.typ {
// 	case indirEngine:
// 		if v.IsNil() {
// 			enc.b.Write(encNil)
// 			return
// 		}
// 		elem := engine.elems[0]
// 		if elem.raw {
// 			enc.encode(v.Elem().Interface(), true)
// 		} else if elem.op > 0 {
// 			encOps[elem.op](enc, v.Elem(), true)
// 		} else {
// 			enc.runEngine(elem.engine, v.Elem())
// 		}
// 	case dictEngine:
// 		elem := engine.elems[0]
// 		enc.b.Write(encDict)
// 		enc.b.Write(elem.typ)
// 		// if enc.raw {
// 		// 	enc.writeDict(v.Interface().(map[string]interface{}), enc.b)
// 		// 	return
// 		// }
// 		// if v.IsNil() {
// 		// 	enc.b.Write(encNil)
// 		// 	return
// 		// }
// 		// enc.writeSize(v.Len(), enc.b)
// 		// if elem.raw {
// 		// case int64:

// 		// }
// 	case mapEngine:
// 		key, elem := engine.elems[0], engine.elems[1]
// 		enc.b.Write(encMap)
// 		enc.b.Write(key.typ)
// 		enc.b.Write(elem.typ)
// 		if v.IsNil() {
// 			enc.b.Write(encNil)
// 			return
// 		}
// 		enc.writeSize(v.Len(), enc.b)

func (enc *Encoder) writeByteSlice(value []byte, buf *bytes.Buffer) {
	if value == nil {
		buf.Write(encNil)
		return
	}
	enc.writeSize(len(value), buf)
	buf.Write(value)
}

func (enc *Encoder) writeDict(value map[string]interface{}, buf *bytes.Buffer) {
	if value == nil {
		buf.Write(encNil)
		return
	}
	enc.writeSize(len(value), buf)
	for k, v := range value {
		enc.writeSize(len(k), buf)
		buf.Write([]byte(k))
		enc.encode(v)
	}
}

func (enc *Encoder) writeFloat32(value float32, buf *bytes.Buffer) {
	v := math.Float32bits(value)
	data := enc.scratch
	data[0] = byte(v >> 24)
	data[1] = byte(v >> 16)
	data[2] = byte(v >> 8)
	data[3] = byte(v)
	buf.Write(data[:4])
}

func (enc *Encoder) writeFloat64(value float64, buf *bytes.Buffer) {
	v := math.Float64bits(value)
	data := enc.scratch
	data[0] = byte(v >> 56)
	data[1] = byte(v >> 48)
	data[2] = byte(v >> 40)
	data[3] = byte(v >> 32)
	data[4] = byte(v >> 24)
	data[5] = byte(v >> 16)
	data[6] = byte(v >> 8)
	data[7] = byte(v)
	buf.Write(data[:8])
}

func (enc *Encoder) writeInt64(value int64, buf *bytes.Buffer) {
	var x uint64
	if value < 0 {
		x = uint64(^value<<1) | 1
	} else {
		x = uint64(value << 1)
	}
	enc.writeUint64(x, buf)
}

func (enc *Encoder) writeSize(value int, buf *bytes.Buffer) {
	i := 0
	for {
		left := value & 127
		value >>= 7
		if value > 0 {
			left += 128
		}
		enc.scratch[i] = byte(left)
		i += 1
		if value == 0 {
			break
		}
	}
	buf.Write(enc.scratch[:i])
}

func (enc *Encoder) writeSlice(value []interface{}, buf *bytes.Buffer) {
	if value == nil {
		buf.Write(encNil)
		return
	}
	enc.writeSize(len(value), buf)
	for _, item := range value {
		enc.encode(item)
	}
}

func (enc *Encoder) writeStringSlice(value []string, buf *bytes.Buffer) {
	if value == nil {
		buf.Write(encNil)
		return
	}
	enc.writeSize(len(value), buf)
	for _, item := range value {
		enc.writeSize(len(item), buf)
		buf.Write([]byte(item))
	}
}

func (enc *Encoder) writeStructInfo(value *encEngine, buf *bytes.Buffer) {
	buf.Write(encStructInfo)
	enc.b.Write(value.typeId[1:])
	enc.writeSize(value.i, buf)
	for _, f := range value.elems {
		enc.writeSize(len(f.name), buf)
		buf.Write(f.name)
		buf.Write(f.engine.typeId)
	}
}

func (enc *Encoder) writeUint64(value uint64, buf *bytes.Buffer) {
	i := 0
	for {
		left := value & 127
		value >>= 7
		if value > 0 {
			left += 128
		}
		enc.scratch[i] = byte(left)
		i += 1
		if value == 0 {
			break
		}
	}
	buf.Write(enc.scratch[:i])
}

func EncodeBigDecimal(value *big.Decimal) []byte {
	left, right := value.Components()
	encoding := encodeBigInt(left, bigintMagicNumber1)
	if right != nil {
		if left.IsNegative() {
			encoding = append(encoding, '\xff')
		} else {
			encoding = append(encoding, '\x00')
		}
		encoding = append(encoding, encodeBigInt(right, bigintMagicNumber2)...)
	}
	return encoding
}

func EncodeBigInt(value *big.Int) []byte {
	return encodeBigInt(value, bigintMagicNumber1)
}

func encodeBigInt(value *big.Int, cutoff *big.Int) []byte {
	if value.IsZero() {
		return []byte{'\x80', '\x01', '\x01'}
	}
	if !value.IsNegative() {
		if value.Cmp(cutoff) == -1 {
			encoding := []byte{'\x80', '\x01', '\x01'}
			mod := big.NewInt(0)
			div, mod := value.DivMod(value, bigint255, mod)
			encoding[2] = byte(mod.Int64()) + 1
			if div.Sign() == 1 {
				div, mod = div.DivMod(div, bigint255, mod)
				encoding[1] = byte(mod.Int64()) + 1
				if div.Sign() == 1 {
					encoding[0] = byte(div.Int64()) + 128
				}
			}
			return encoding
		}
		value = value.Sub(value, cutoff)
		encoding := []byte{'\xff'}
		left := big.NewInt(0)
		lead, left := value.DivMod(value, bigint255, left)
		var n int64 = 1
		exp := big.NewInt(0)
		div := big.NewInt(0)
		for (div.Div(lead, exp.Exp(big.NewInt(253), big.NewInt(n), nil))).Sign() == 1 {
			n += 1
		}
		encoding = append(encoding, byte(n)+1, '\xff')
		leadChars := make([]byte, 0)
		mod := big.NewInt(0)
		for {
			if lead.IsZero() {
				break
			}
			lead, mod = lead.DivMod(lead, bigint253, mod)
			leadChars = append(leadChars, byte(mod.Int64())+2)
		}
		lenLead := len(leadChars)
		if lenLead > 0 {
			for i := lenLead - 1; i >= 0; i-- {
				encoding = append(encoding, leadChars[i])
			}
		}
		if left.Sign() == 1 {
			encoding = append(encoding, '\x01', byte(left.Int64()))
		}
		return encoding
	}
	value = value.Neg(value)
	if value.Cmp(cutoff) == -1 {
		encoding := []byte{'\x7f', '\xfe', '\xfe'}
		mod := big.NewInt(0)
		div, mod := value.DivMod(value, bigint255, mod)
		encoding[2] = 254 - byte(mod.Int64())
		if div.Sign() == 1 {
			div, mod = div.DivMod(div, bigint255, mod)
			encoding[1] = 254 - byte(mod.Int64())
			if div.Sign() == 1 {
				encoding[0] = 127 - byte(div.Int64())
			}
		}
		return encoding
	}
	value = value.Sub(value, cutoff)
	encoding := []byte{'\x00'}
	left := big.NewInt(0)
	lead, left := value.DivMod(value, bigint254, left)
	var n int64 = 1
	exp := big.NewInt(0)
	div := big.NewInt(0)
	for (div.Div(lead, exp.Exp(big.NewInt(253), big.NewInt(n), nil))).Sign() == 1 {
		n += 1
	}
	encoding = append(encoding, 254-byte(n), '\x00')
	leadChars := make([]byte, 0)
	mod := big.NewInt(0)
	for {
		if lead.IsZero() {
			break
		}
		lead, mod = lead.DivMod(lead, bigint253, mod)
		leadChars = append(leadChars, byte(253-mod.Int64()))
	}
	lenLead := len(leadChars)
	if lenLead > 0 {
		for i := lenLead - 1; i >= 0; i-- {
			encoding = append(encoding, leadChars[i])
		}
	}
	if lenLead > 1 {
		encoding = append(encoding, '\x00')
	}
	encoding = append(encoding, '\xfe')
	if left.Sign() == 1 {
		encoding = append(encoding, '\x01', 254-byte(left.Int64()))
	} else {
		encoding = append(encoding, '\xfe')
	}
	return encoding
}

func EncodeInt64(value int64) []byte {
	var x uint64
	if value < 0 {
		x = uint64(^value<<1) | 1
	} else {
		x = uint64(value << 1)
	}
	return EncodeUint64(x)
}

func EncodeInt64AsBig(value int64) []byte {
	if value == 0 {
		return []byte{'\x80', '\x01', '\x01'}
	}
	if value > 0 {
		if value < magicNumber {
			encoding := []byte{'\x80', '\x01', '\x01'}
			div, mod := value/255, value%255
			encoding[2] = byte(mod) + 1
			if div > 0 {
				div, mod = div/255, div%255
				encoding[1] = byte(mod) + 1
				if div > 0 {
					encoding[0] = byte(div) + 128
				}
			}
			return encoding
		}
		value -= magicNumber
		encoding := []byte{'\xff'}
		lead, left := value/255, value%255
		var n int64 = 1
		for (lead / pow(253, n)) > 0 {
			n += 1
		}
		encoding = append(encoding, byte(n)+1, '\xff')
		leadChars := make([]byte, 0)
		for {
			var mod int64
			if lead == 0 {
				break
			}
			lead, mod = lead/253, lead%253
			leadChars = append(leadChars, byte(mod)+2)
		}
		lenLead := len(leadChars)
		if lenLead > 0 {
			for i := lenLead - 1; i >= 0; i-- {
				encoding = append(encoding, leadChars[i])
			}
		}
		if left > 0 {
			encoding = append(encoding, '\x01', byte(left))
		}
		return encoding
	}
	value = -value
	if value < magicNumber {
		encoding := []byte{'\x7f', '\xfe', '\xfe'}
		div, mod := value/255, value%255
		encoding[2] = 254 - byte(mod)
		if div > 0 {
			div, mod = div/255, div%255
			encoding[1] = 254 - byte(mod)
			if div > 0 {
				encoding[0] = 127 - byte(div)
			}
		}
		return encoding
	}
	value -= magicNumber
	encoding := []byte{'\x00'}
	lead, left := value/254, value%254
	var n int64 = 1
	for (lead / pow(253, n)) > 0 {
		n += 1
	}
	encoding = append(encoding, 254-byte(n), '\x00')
	leadChars := make([]byte, 0)
	for {
		var mod int64
		if lead == 0 {
			break
		}
		lead, mod = lead/253, lead%253
		leadChars = append(leadChars, 253-byte(mod))
	}
	lenLead := len(leadChars)
	if lenLead > 0 {
		for i := lenLead - 1; i >= 0; i-- {
			encoding = append(encoding, leadChars[i])
		}
	}
	if lenLead > 1 {
		encoding = append(encoding, '\x00')
	}
	encoding = append(encoding, '\xfe')
	if left > 0 {
		encoding = append(encoding, 254-byte(left))
	} else {
		encoding = append(encoding, '\xfe')
	}
	return encoding
}

func EncodeSize(value int) ([]byte, os.Error) {
	if value < 0 || value > maxInt32 {
		return nil, OutOfRangeError
	}
	data := make([]byte, 6)
	i := 0
	for {
		left := value & 127
		value >>= 7
		if value > 0 {
			left += 128
		}
		data[i] = byte(left)
		i += 1
		if value == 0 {
			break
		}
	}
	return data[:i], nil
}

func EncodeStringNumber(value string) ([]byte, os.Error) {
	if strings.Count(value, ".") > 0 {
		number, ok := big.NewDecimal(value)
		if !ok {
			return nil, Error("couldn't create a big.Decimal representation of " + value)
		}
		return EncodeBigDecimal(number), nil
	}
	number, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return nil, Error("couldn't create an big.Int representation of " + value)
	}
	return EncodeBigInt(number), nil
}

func EncodeUint64(value uint64) []byte {
	data := make([]byte, 11)
	i := 0
	for {
		left := value & 127
		value >>= 7
		if value > 0 {
			left += 128
		}
		data[i] = byte(left)
		i += 1
		if value == 0 {
			break
		}
	}
	return data[:i]
}

func pow(x, y int64) (z int64) {
	var i int64
	z = 1
	for i = 0; i < y; i++ {
		z = z * x
	}
	return z
}

func compileEncEngine(rt reflect.Type) (*encEngine, os.Error) {

	if engine, present := encEngines[rt]; present {
		return engine, nil
	}

	engine := &encEngine{}
	kind := rt.Kind()

	if op := encBaseOps[kind]; op > 0 {
		engine.typeId = encTypeIndicator[op]
		engine.i = op
		return engine, nil
	}

	switch rt.Kind() {
	case reflect.Slice:
		et := rt.Elem()
		ekind := et.Kind()
		if op := encSliceOps[ekind]; op > 0 {
			engine.i = op
			engine.typeId = encTypeIndicator[op]
			return engine, nil
		}
		if ekind == reflect.Slice {
			if op := encSliceSliceOps[et.Elem().Kind()]; op > 0 {
				engine.i = op
				engine.typeId = encTypeIndicator[op]
				return engine, nil
			}
		}
		if et == anyInterfaceType {
			engine.i = SliceAny
			engine.typeId = encSliceAny
			return engine, nil
		}
		elemEngine, err := compileEncEngine(et)
		if err != nil {
			return nil, err
		}
		switch elemEngine.typ {
		default:
			engine.typeId = []byte{Slice, Any}
		case indirEngine:
			if elemEngine.explicit {
				engine.typeId = []byte{Slice, Any}
			} else {
				engine.typeId = append([]byte{Slice}, elemEngine.typeId...)
			}
		case rawEngine:
			engine.typeId = []byte{Slice, byte(elemEngine.i)}
		case structEngine:
			engine.typeId = append([]byte{Slice}, elemEngine.typeId...)
		}
		engine.elems = []*encElem{&encElem{engine: elemEngine, offset: et.Size()}}
		engine.explicit = true
		engine.typ = sliceEngine
	case reflect.Interface:
		engine.explicit = true
		engine.typ = reflectEngine
		engine.typeId = encAny
	case reflect.Ptr:
		i := 1
		for {
			rt = rt.Elem()
			if rt.Kind() != reflect.Ptr {
				break
			}
			i++
		}
		elemEngine, err := compileEncEngine(rt)
		if err != nil {
			return nil, err
		}
		if elemEngine.typ == rawEngine || elemEngine.typ == structEngine {
			engine.explicit = false
		} else {
			engine.explicit = true
		}
		engine.elems = []*encElem{&encElem{engine: elemEngine}}
		engine.i = i
		engine.typ = indirEngine
		engine.typeId = elemEngine.typeId
	// case reflect.Map:
	// 	engine.elems = make([]*encElem, 2)
	// 	engine.typ = mapEngine
	case reflect.Struct:
		engine.elems = make([]*encElem, 0)
		n := 0
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			if field.Anonymous {
				continue
			}
			// TODO(tav): can this happen?
			// if field.Name == "" {
			// 	continue
			// }
			var name string
			if tag := field.Tag.Get("argo"); tag != "" {
				if tag == "-" {
					continue
				}
				name = tag
			}
			if name == "" {
				name = field.Name
				rune, _ := utf8.DecodeRuneInString(name)
				if !unicode.IsUpper(rune) {
					continue
				}
			}
			elemEngine, err := compileEncEngine(field.Type)
			if err != nil {
				return nil, err
			}
			elem := &encElem{engine: elemEngine, name: []byte(name), idx: i, offset: field.Offset}
			engine.elems = append(engine.elems, elem)
			n += 1
		}
		engine.i = n
		engine.rt = rt
		engine.typ = structEngine
		engine.typeId = append([]byte{Struct}, EncodeUint64(nextEncId)...)
		nextEncId += 1
	default:
		error(Error("unknown type to compile an encoding engine: " + rt.String()))
	}

	encEngines[rt] = engine
	return engine, nil

}

func NewEncoder(w io.Writer) *Encoder {
	if b, ok := w.(*bytes.Buffer); ok {
		return &Encoder{
			b:       b,
			dup:     false,
			scratch: make([]byte, 11),
			w:       w,
		}
	}
	return &Encoder{
		b:       &bytes.Buffer{},
		dup:     true,
		scratch: make([]byte, 11),
		w:       w,
	}
}

func unsafeAddr(v reflect.Value) uintptr {
	if v.CanAddr() {
		return v.UnsafeAddr()
	}
	x := reflect.New(v.Type()).Elem()
	x.Set(v)
	return x.UnsafeAddr()
}

var nextEncId uint64 = baseId

func init() {
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
