// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	"amp/big"
	"bytes"
	"io"
	"math"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
	"unsafe"
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

var (
	anyInterfaceType = reflect.TypeOf(interface{}(nil))
	encEngines       = map[reflect.Type]*encEngine{}
	nextEncId        = uint64(baseId)
)

var (
	encCompiler sync.RWMutex
	is64bit     bool
)

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
	reflect.String: StringSlice,
	reflect.Uint8:  ByteSlice,
}

var (
	encAny         = []byte{Any}
	encBigDecimal  = []byte{BigDecimal}
	encBigInt      = []byte{BigInt}
	encBool        = []byte{Bool}
	encByte        = []byte{Byte}
	encByteSlice   = []byte{ByteSlice}
	encComplex64   = []byte{Complex64}
	encComplex128  = []byte{Complex128}
	encDict        = []byte{Dict}
	encFloat32     = []byte{Float32}
	encFloat64     = []byte{Float64}
	encInt32       = []byte{Int32}
	encInt64       = []byte{Int64}
	encItem        = []byte{Item}
	encMap         = []byte{Map}
	encNil         = []byte{Nil}
	encSlice       = []byte{Slice}
	encString      = []byte{String}
	encStringSlice = []byte{StringSlice}
	encStruct      = []byte{Struct}
	encStructInfo  = []byte{StructInfo}
	encUint32      = []byte{Uint32}
	encUint64      = []byte{Uint64}
)

var (
	encDictAny   = []byte{Dict, Any}
	encSliceAny  = []byte{Slice, Any}
	encTrue      = []byte{1}
	encFalse     = []byte{0}
	encBoolTrue  = []byte{Bool, 1}
	encBoolFalse = []byte{Bool, 0}
)

var encTypeIndicator = [...][]byte{
	Any:         encAny,
	BigDecimal:  encBigDecimal,
	BigInt:      encBigInt,
	Bool:        encBool,
	Byte:        encByte,
	ByteSlice:   encByteSlice,
	Complex64:   encComplex64,
	Complex128:  encComplex128,
	Dict:        encDict,
	Float32:     encFloat32,
	Float64:     encFloat64,
	Int32:       encInt32,
	Int64:       encInt64,
	Item:        encItem,
	Map:         encMap,
	Nil:         encNil,
	Slice:       encSlice,
	String:      encString,
	StringSlice: encStringSlice,
	Struct:      encStruct,
	StructInfo:  encStructInfo,
	Uint32:      encUint32,
	Uint64:      encUint64,
}

type encEngine struct {
	elems  []*encElem
	i      int /* array: length, indir: indir, raw: op, struct: length */
	rt     reflect.Type
	runner func(*Encoder, *encEngine, uintptr)
	typ    int
	typeId []byte
}

type encElem struct {
	engine *encEngine
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
				raise(PointerError)
			}
			indir--
		}
		elemEngine := engine.elems[0].engine
		elemEngine.runner(enc, elemEngine, uintptr(up))
	},
	reflectEngine: func(enc *Encoder, engine *encEngine, p uintptr) {
		enc.encode(unsafe.Unreflect(engine.rt, unsafe.Pointer(p)))
	},
	sliceEngine: func(enc *Encoder, engine *encEngine, p uintptr) {
		enc.b.Write(engine.typeId)
		header := (*reflect.SliceHeader)(unsafe.Pointer(p))
		mark := header.Data
		offset := engine.elems[0].offset
		elemEngine := engine.elems[0].engine
		if elemEngine.typ == indirEngine {
			subEngine := elemEngine.elems[0].engine
			if subEngine.typ == structEngine && header.Len > 100 {
				fieldElems := subEngine.elems
				engines := make([]*encEngine, len(fieldElems))
				offsets := make([]uintptr, len(fieldElems))
				for i, field := range fieldElems {
					engines[i] = field.engine
					offsets[i] = field.offset
				}
				indir := elemEngine.i
				for i := 0; i < header.Len; i++ {
					sp := uintptr(encIndirect(indir, mark))
					for j, field := range engines {
						field.runner(enc, field, sp+offsets[j])
					}
					mark += offset
				}
				return
			}
			runner := subEngine.runner
			indir := elemEngine.i
			for i := 0; i < header.Len; i++ {
				runner(enc, subEngine, uintptr(encIndirect(indir, mark)))
				mark += offset
			}
			return
		}
		runner := elemEngine.runner
		for i := 0; i < header.Len; i++ {
			runner(enc, elemEngine, mark)
			mark += offset
		}
	},
	structEngine: func(enc *Encoder, engine *encEngine, p uintptr) {
		for _, elem := range engine.elems {
			field := elem.engine
			field.runner(enc, field, p+elem.offset)
		}
	},
}

var encOps = [...]func(*Encoder, *encEngine, uintptr){
	Bool: func(enc *Encoder, engine *encEngine, p uintptr) {
		v := *(*bool)(unsafe.Pointer(p))
		if v {
			enc.b.Write(encTrue)
		} else {
			enc.b.Write(encFalse)
		}
	},
	Byte: func(enc *Encoder, engine *encEngine, p uintptr) {
		v := *(*byte)(unsafe.Pointer(p))
		enc.b.Write([]byte{v})
	},
	ByteSlice: func(enc *Encoder, engine *encEngine, p uintptr) {
		v := *(*[]byte)(unsafe.Pointer(p))
		if v == nil {
			enc.b.Write(encNil)
			return
		}
		enc.writeByteSlice(v)
	},
	Complex64: func(enc *Encoder, engine *encEngine, p uintptr) {
		v := *(*complex64)(unsafe.Pointer(p))
		enc.writeFloat32(float32(real(v)))
		enc.writeFloat32(float32(imag(v)))
	},
	Complex128: func(enc *Encoder, engine *encEngine, p uintptr) {
		v := *(*complex128)(unsafe.Pointer(p))
		enc.writeFloat64(real(v))
		enc.writeFloat64(imag(v))
	},
	Float32: func(enc *Encoder, engine *encEngine, p uintptr) {
		v := *(*float32)(unsafe.Pointer(p))
		enc.writeFloat32(v)
	},
	Float64: func(enc *Encoder, engine *encEngine, p uintptr) {
		v := *(*float64)(unsafe.Pointer(p))
		enc.writeFloat64(v)
	},
	Int32: func(enc *Encoder, engine *encEngine, p uintptr) {
		return
		v := *(*int32)(unsafe.Pointer(p))
		enc.writeInt64(int64(v))
	},
	Int64: func(enc *Encoder, engine *encEngine, p uintptr) {
		v := *(*int64)(unsafe.Pointer(p))
		enc.writeInt64(v)
	},
	String: func(enc *Encoder, engine *encEngine, p uintptr) {
		v := *(*string)(unsafe.Pointer(p))
		enc.writeSize(len(v))
		enc.b.Write([]byte(v))
	},
	StringSlice: func(enc *Encoder, engine *encEngine, p uintptr) {
		v := *(*[]string)(unsafe.Pointer(p))
		enc.writeStringSlice(v, enc.b)
	},
	Uint32: func(enc *Encoder, engine *encEngine, p uintptr) {
		v := *(*uint32)(unsafe.Pointer(p))
		enc.writeUint64(uint64(v))
	},
	Uint64: func(enc *Encoder, engine *encEngine, p uintptr) {
		v := *(*uint64)(unsafe.Pointer(p))
		enc.writeUint64(v)
	},
}

func encDictAnyOp(enc *Encoder, engine *encEngine, p uintptr) {
	v := *(*map[string]interface{})(unsafe.Pointer(p))
	if v == nil {
		enc.b.Write(encNil)
		return
	}
	enc.writeSize(len(v))
	for k, elem := range v {
		enc.writeSize(len(k))
		enc.b.Write([]byte(k))
		enc.encode(elem)
	}
}

func encSliceAnyOp(enc *Encoder, engine *encEngine, p uintptr) {
	v := *(*[]interface{})(unsafe.Pointer(p))
	if v == nil {
		enc.b.Write(encNil)
		return
	}
	enc.writeSize(len(v))
	for _, elem := range v {
		enc.encode(elem)
	}
}

type Encoder struct {
	b       *bytes.Buffer
	dup     bool
	engines map[reflect.Type]*encEngine
	scratch []byte
	w       io.Writer
}

func (enc *Encoder) Encode(v interface{}) (err error) {
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
		enc.writeSize(len(val))
		enc.b.Write(val)
	case *big.Int:
		enc.b.Write(encBigInt)
		val := EncodeBigInt(value)
		enc.writeSize(len(val))
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
		enc.writeByteSlice(value)
	case complex64:
		enc.b.Write(encComplex64)
		enc.writeFloat32(real(value))
		enc.writeFloat32(imag(value))
	case complex128:
		enc.b.Write(encComplex128)
		enc.writeFloat64(real(value))
		enc.writeFloat64(imag(value))
	case float32:
		enc.b.Write(encFloat32)
		enc.writeFloat32(value)
	case float64:
		enc.b.Write(encFloat64)
		enc.writeFloat64(value)
	case int:
		if is64bit {
			enc.b.Write(encInt64)
		} else {
			enc.b.Write(encInt32)
		}
		enc.writeInt64(int64(value))
	case int8:
		enc.b.Write(encInt32)
		enc.writeInt64(int64(value))
	case int16:
		enc.b.Write(encInt32)
		enc.writeInt64(int64(value))
	case int32:
		enc.b.Write(encInt32)
		enc.writeInt64(int64(value))
	case int64:
		enc.b.Write(encInt64)
		enc.writeInt64(value)
	case []interface{}:
		enc.b.Write(encSliceAny)
		enc.writeSlice(value, enc.b)
	case map[string]interface{}:
		enc.b.Write(encDictAny)
		enc.writeDict(value, enc.b)
	case string:
		enc.b.Write(encString)
		enc.writeSize(len(value))
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
		enc.writeUint64(uint64(value))
	case uint16:
		enc.b.Write(encUint32)
		enc.writeUint64(uint64(value))
	case uint32:
		enc.b.Write(encUint32)
		enc.writeUint64(uint64(value))
	case uint64:
		enc.b.Write(encUint64)
		enc.writeUint64(value)
	default:
		enc.encodeValue(reflect.ValueOf(v))
	}
}

func (enc *Encoder) encodeValue(rv reflect.Value) {

	if !rv.IsValid() {
		raise(Error("invalid value detected"))
	}

	rt := rv.Type()
	if engine, present := enc.engines[rt]; present {
		enc.b.Write(engine.typeId)
		engine.runner(enc, engine, unsafeAddr(rv))
		return
	}

	encCompiler.Lock()
	engine, err := compileEncEngine(rt)
	encCompiler.Unlock()

	if err != nil {
		raise(err)
	}

	enc.checkinEngine(engine)
	enc.b.Write(engine.typeId)
	engine.runner(enc, engine, unsafeAddr(rv))

}

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

func (enc *Encoder) writeByteSlice(value []byte) {
	if value == nil {
		enc.b.Write(encNil)
		return
	}
	enc.writeSize(len(value))
	enc.b.Write(value)
}

func (enc *Encoder) writeDict(value map[string]interface{}, buf *bytes.Buffer) {
	if value == nil {
		buf.Write(encNil)
		return
	}
	enc.writeSize(len(value))
	for k, v := range value {
		enc.writeSize(len(k))
		buf.Write([]byte(k))
		enc.encode(v)
	}
}

func (enc *Encoder) writeFloat32(value float32) {
	v := math.Float32bits(value)
	data := enc.scratch
	data[0] = byte(v >> 24)
	data[1] = byte(v >> 16)
	data[2] = byte(v >> 8)
	data[3] = byte(v)
	enc.b.Write(data[:4])
}

func (enc *Encoder) writeFloat64(value float64) {
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
	enc.b.Write(data[:8])
}

func (enc *Encoder) writeInt64(value int64) {
	var x uint64
	if value < 0 {
		x = uint64(^value<<1) | 1
	} else {
		x = uint64(value << 1)
	}
	enc.writeUint64(x)
}

func (enc *Encoder) writeSize(value int) {
	i := 0
	for value >= 128 {
		enc.scratch[i] = byte(value) | 128
		value >>= 7
		i++
	}
	enc.scratch[i] = byte(value)
	enc.b.Write(enc.scratch[:i+1])
}

func (enc *Encoder) writeSlice(value []interface{}, buf *bytes.Buffer) {
	if value == nil {
		buf.Write(encNil)
		return
	}
	enc.writeSize(len(value))
	for _, item := range value {
		enc.encode(item)
	}
}

func (enc *Encoder) writeStringSlice(value []string, buf *bytes.Buffer) {
	if value == nil {
		buf.Write(encNil)
		return
	}
	enc.writeSize(len(value))
	for _, item := range value {
		enc.writeSize(len(item))
		buf.Write([]byte(item))
	}
}

func (enc *Encoder) writeStructInfo(value *encEngine, buf *bytes.Buffer) {
	buf.Write(encStructInfo)
	buf.Write(value.typeId[1:])
	enc.writeSize(value.i)
	for _, f := range value.elems {
		enc.writeSize(len(f.name))
		buf.Write(f.name)
		buf.Write(f.engine.typeId)
	}
}

func (enc *Encoder) writeUint64(value uint64) {
	i := 0
	for value >= 128 {
		enc.scratch[i] = byte(value) | 128
		value >>= 7
		i++
	}
	enc.scratch[i] = byte(value)
	enc.b.Write(enc.scratch[:i+1])
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

func EncodeSize(value int) ([]byte, error) {
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

func EncodeStringNumber(value string) ([]byte, error) {
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

func compileEncEngine(rt reflect.Type) (*encEngine, error) {

	if engine, present := encEngines[rt]; present {
		return engine, nil
	}

	engine := &encEngine{}
	kind := rt.Kind()

	if op := encBaseOps[kind]; op > 0 {
		engine.runner = encOps[op]
		engine.typeId = encTypeIndicator[op]
		return engine, nil
	}

	switch rt.Kind() {
	case reflect.Slice:
		et := rt.Elem()
		if op := encSliceOps[et.Kind()]; op > 0 {
			engine.runner = encOps[op]
			engine.typeId = encTypeIndicator[op]
			return engine, nil
		}
		if et == anyInterfaceType {
			engine.runner = encSliceAnyOp
			engine.typeId = encSliceAny
			return engine, nil
		}
		elemEngine, err := compileEncEngine(et)
		if err != nil {
			return nil, err
		}
		if elemEngine.typ == indirEngine && elemEngine.elems[0].engine.typ == rawEngine {
			switch elemEngine.typeId[0] {
			case Byte:
				engine.typeId = encByteSlice
			case String:
				engine.typeId = encStringSlice
			default:
				engine.typeId = append([]byte{Slice}, elemEngine.typeId...)
			}
		} else {
			engine.typeId = append([]byte{Slice}, elemEngine.typeId...)
		}
		engine.elems = []*encElem{&encElem{engine: elemEngine, offset: et.Size()}}
		engine.runner = encEngineRunners[sliceEngine]
		engine.typ = sliceEngine
	case reflect.Interface:
		engine.rt = rt
		engine.runner = encEngineRunners[reflectEngine]
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
		engine.elems = []*encElem{&encElem{engine: elemEngine}}
		engine.i = i
		engine.runner = encEngineRunners[indirEngine]
		engine.typ = indirEngine
		engine.typeId = elemEngine.typeId
	case reflect.Map:
		kt := rt.Key()
		et := rt.Elem()
		if kt.Kind() == reflect.String {
			if et.Kind() == reflect.Interface {
				engine.runner = encDictAnyOp
				engine.typeId = encDictAny
				return engine, nil
			}
		}
		return nil, Error("unsupported encoding engine type: " + rt.String())
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
			elem := &encElem{engine: elemEngine, name: []byte(name), offset: field.Offset}
			engine.elems = append(engine.elems, elem)
			n += 1
		}
		engine.i = n
		engine.rt = rt
		engine.runner = encEngineRunners[structEngine]
		engine.typ = structEngine
		engine.typeId = append([]byte{Struct}, EncodeUint64(nextEncId)...)
		nextEncId += 1
	default:
		return nil, Error("unsupported encoding engine type: " + rt.String())
	}

	encEngines[rt] = engine
	return engine, nil

}

func encIndirect(indir int, p uintptr) unsafe.Pointer {
	up := unsafe.Pointer(p)
	for indir > 0 {
		up = *(*unsafe.Pointer)(up)
		if up == nil {
			raise(PointerError)
		}
		indir--
	}
	return up
}

func pow(x, y int64) (z int64) {
	var i int64
	z = 1
	for i = 0; i < y; i++ {
		z = z * x
	}
	return z
}

func NewEncoder(w io.Writer) *Encoder {
	if b, ok := w.(*bytes.Buffer); ok {
		return &Encoder{
			b:       b,
			dup:     false,
			engines: make(map[reflect.Type]*encEngine),
			scratch: make([]byte, 11),
			w:       w,
		}
	}
	return &Encoder{
		b:       &bytes.Buffer{},
		dup:     true,
		engines: make(map[reflect.Type]*encEngine),
		scratch: make([]byte, 11),
		w:       w,
	}
}
