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
	"utf8"
)

const (
	baseId   = 64
	maxInt32 = 2147483647
)

const (
	indirEngine = iota
	mapEngine
	sliceEngine
	structEngine
)

var (
	encAny            = []byte{Any}
	encBigDecimal     = []byte{BigDecimal}
	encBigInt         = []byte{BigInt}
	encBool           = []byte{Bool}
	encBoolFalse      = []byte{BoolFalse}
	encBoolTrue       = []byte{BoolTrue}
	encByte           = []byte{Byte}
	encByteSlice      = []byte{ByteSlice}
	encByteSliceSlice = []byte{ByteSliceSlice}
	encComplex64      = []byte{Complex64}
	encComplex128     = []byte{Complex128}
	encDict           = []byte{Dict}
	encDictAny        = []byte{Dict, Any}
	encFloat32        = []byte{Float32}
	encFloat64        = []byte{Float64}
	encHeader         = []byte{Header}
	encInt32          = []byte{Int32}
	encInt64          = []byte{Int64}
	encItem           = []byte{Item}
	encMap            = []byte{Map}
	encNil            = []byte{Nil}
	encSlice          = []byte{Slice}
	encSliceAny       = []byte{Slice, Any}
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
	reflect.Int:        Int64,
	reflect.Int8:       Int32,
	reflect.Int16:      Int32,
	reflect.Int32:      Int32,
	reflect.Int64:      Int64,
	reflect.String:     String,
	reflect.Uint:       Uint64,
	reflect.Uint8:      Byte,
	reflect.Uint16:     Uint32,
	reflect.Uint32:     Uint32,
	reflect.Uint64:     Uint64,
}

var (
	anyInterfaceType   = reflect.TypeOf(interface{}(nil))
	byteSliceType      = reflect.TypeOf([]byte(nil))
	byteSliceSliceType = reflect.TypeOf([][]byte(nil))
	stringSliceType    = reflect.TypeOf([]string(nil))
)

var encRawTypes = map[reflect.Type]bool{
	byteSliceType:      true,
	byteSliceSliceType: true,
	stringSliceType:    true,
}

var encOps = [...]func(*Encoder, reflect.Value, bool){
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
		enc.writeString(v.String(), enc.b)
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
	pad     []byte
	err     os.Error
	seen    map[*structInfo]bool
	w       io.Writer
}

func (enc *Encoder) Encode(v interface{}) (err os.Error) {
	enc.encode(v, true)
	if enc.err != nil {
		return enc.err
	}
	if enc.dup {
		_, err = enc.w.Write(enc.b.Bytes())
	}
	return
}

func (enc *Encoder) encode(v interface{}, typeinfo bool) {
	switch value := v.(type) {
	case *big.Decimal:
		if typeinfo {
			enc.b.Write(encBigDecimal)
		}
		WriteBigDecimal(value, enc.b)
	case *big.Int:
		if typeinfo {
			enc.b.Write(encBigInt)
		}
		WriteBigInt(value, enc.b)
	case bool:
		if value {
			enc.b.Write(encBoolTrue)
		} else {
			enc.b.Write(encBoolFalse)
		}
	case byte:
		if typeinfo {
			enc.b.Write(encByte)
		}
		enc.b.Write([]byte{value})
	case []byte:
		if typeinfo {
			enc.b.Write(encByteSlice)
		}
		enc.writeByteSlice(value, enc.b)
	case complex64:
		if typeinfo {
			enc.b.Write(encComplex64)
		}
		enc.writeFloat32(real(value), enc.b)
		enc.writeFloat32(imag(value), enc.b)
	case complex128:
		if typeinfo {
			enc.b.Write(encComplex128)
		}
		enc.writeFloat64(real(value), enc.b)
		enc.writeFloat64(imag(value), enc.b)
	case float32:
		if typeinfo {
			enc.b.Write(encFloat32)
		}
		enc.writeFloat32(value, enc.b)
	case float64:
		if typeinfo {
			enc.b.Write(encFloat64)
		}
		enc.writeFloat64(value, enc.b)
	case int:
		if typeinfo {
			enc.b.Write(encInt64)
		}
		enc.writeInt64(int64(value), enc.b)
	case int8:
		if typeinfo {
			enc.b.Write(encInt32)
		}
		enc.writeInt64(int64(value), enc.b)
	case int16:
		if typeinfo {
			enc.b.Write(encInt32)
		}
		enc.writeInt64(int64(value), enc.b)
	case int32:
		if typeinfo {
			enc.b.Write(encInt32)
		}
		enc.writeInt64(int64(value), enc.b)
	case int64:
		if typeinfo {
			enc.b.Write(encInt64)
		}
		enc.writeInt64(value, enc.b)
	case []interface{}:
		if typeinfo {
			enc.b.Write(encSliceAny)
		}
		enc.writeSlice(value, enc.b)
	case map[string]interface{}:
		if typeinfo {
			enc.b.Write(encDictAny)
		}
		enc.writeDict(value, enc.b)
	case rpc.Header:
		if typeinfo {
			enc.b.Write(encHeader)
		}
		enc.writeDict(value, enc.b)
	case string:
		if typeinfo {
			enc.b.Write(encString)
		}
		enc.writeString(value, enc.b)
	case []string:
		if typeinfo {
			enc.b.Write(encStringSlice)
		}
		enc.writeStringSlice(value, enc.b)
	case uint16:
		if typeinfo {
			enc.b.Write(encUint32)
		}
		enc.writeUint64(uint64(value), enc.b)
	case uint32:
		if typeinfo {
			enc.b.Write(encUint32)
		}
		enc.writeUint64(uint64(value), enc.b)
	case uint64:
		if typeinfo {
			enc.b.Write(encUint64)
		}
		enc.writeUint64(value, enc.b)
	default:
		enc.encodeValue(reflect.ValueOf(v), typeinfo)
	}
}

func (enc *Encoder) encodeValue(v reflect.Value, typeinfo bool) {

	if !v.IsValid() {
		error(Error("invalid type detected"))
	}

	if op := encBaseOps[v.Kind()]; op > 0 {
		encOps[op](enc, v, typeinfo)
		return
	}

	if enc.seen == nil {
		enc.engines = make(map[reflect.Type]*encEngine)
		enc.seen = make(map[*structInfo]bool)
	} else if engine, present := enc.engines[v.Type()]; present {
		enc.runEngine(engine, v)
		return
	}

	encCompiler.Lock()
	defer encCompiler.Unlock()

	engine := compileEncEngine(v.Type())
	enc.engines[v.Type()] = engine
	enc.checkinEngine(engine)
	enc.runEngine(engine, v)

}

func (enc *Encoder) checkinEngine(engine *encEngine) {
	if engine.typ == structEngine {
		if !enc.seen[engine.info] {
			enc.writeStructInfo(engine.info, enc.b)
			enc.seen[engine.info] = true
		}
	} else {
		for _, elem := range engine.elems {
			if elem.engine != nil {
				enc.checkinEngine(elem.engine)
			}
		}
	}
}

func (enc *Encoder) runEngine(engine *encEngine, v reflect.Value) {
	switch engine.typ {
	case indirEngine:
		if v.IsNil() {
			enc.b.Write(encNil)
			return
		}
		elem := engine.elems[0]
		if elem.raw {
			enc.encode(v.Elem().Interface(), true)
		} else if elem.op > 0 {
			encOps[elem.op](enc, v.Elem(), true)
		} else {
			enc.runEngine(elem.engine, v.Elem())
		}
	case sliceEngine:
		elem := engine.elems[0]
		enc.b.Write(encSlice)
		enc.b.Write(elem.typ)
		if v.IsNil() {
			enc.b.Write(encNil)
			return
		}
		size := v.Len()
		enc.writeSize(size, enc.b)
		if elem.raw {
			for i := 0; i < size; i++ {
				enc.encode(v.Index(i).Interface(), true)
			}
			return
		}
		if elem.op > 0 {
			op := encOps[elem.op]
			for i := 0; i < size; i++ {
				op(enc, v.Index(i), true)
			}
			return
		}
		for i := 0; i < size; i++ {
			enc.runEngine(elem.engine, v.Index(i))
		}
	case structEngine:
		info := engine.info
		enc.b.Write(encStruct)
		enc.writeUint64(info.id, enc.b)
		for _, f := range info.fields {
			if f.raw {
				enc.encode(v.Field(f.idx).Interface(), f.nested)
			} else if f.op > 0 {
				encOps[f.op](enc, v.Field(f.idx), f.nested)
			} else {
				enc.runEngine(f.engine, v.Field(f.idx))
			}
		}
	}
}

func (enc *Encoder) encodeValue2(v reflect.Value, typeinfo bool) {

	switch v.Kind() {
	case reflect.Map:
		var forcekey bool
		buf := enc.b
		typ := v.Type()
		keytype := typ.Key()
		elemtype := typ.Elem()
		if keytype.Kind() == reflect.String {
			if elemtype == anyInterfaceType {
				buf.Write(encDictAny)
				enc.writeDict(v.Interface().(map[string]interface{}), buf)
				return
			} else {
				buf.Write(encDict)
			}
		} else {
			buf.Write(encMap)
			if keytype == anyInterfaceType {
				buf.Write(encAny)
				forcekey = true
			} else {
				forcekey = enc.writeEncType(keytype)
			}
		}
		var forceval bool
		if elemtype == anyInterfaceType {
			buf.Write(encAny)
			forceval = true
		} else {
			forceval = enc.writeEncType(elemtype)
		}
		if v.IsNil() {
			buf.Write(encNil)
			return
		}
		enc.writeSize(v.Len(), buf)
		for _, key := range v.MapKeys() {
			enc.encode(key.Interface(), forcekey)
			enc.encode(v.MapIndex(key).Interface(), forceval)
		}
	default:
		error(Error("encoding unknown type: " + v.Kind().String()))
	}

}

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
		enc.encode(v, true)
	}
}

func (enc *Encoder) writeEncType(t reflect.Type) bool {
	enctype, nested := getEncType(t)
	if enctype == nil {
		error(Error("unknown element type: " + t.String()))
	}
	enc.b.Write(enctype)
	return nested
}

func (enc *Encoder) writeFloat32(value float32, buf *bytes.Buffer) {
	v := math.Float32bits(value)
	data := enc.pad
	data[0] = byte(v >> 24)
	data[1] = byte(v >> 16)
	data[2] = byte(v >> 8)
	data[3] = byte(v)
	buf.Write(data[:4])
}

func (enc *Encoder) writeFloat64(value float64, buf *bytes.Buffer) {
	v := math.Float64bits(value)
	data := enc.pad
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
		enc.pad[i] = byte(left)
		i += 1
		if value == 0 {
			break
		}
	}
	buf.Write(enc.pad[:i])
}

func (enc *Encoder) writeSlice(value []interface{}, buf *bytes.Buffer) {
	if value == nil {
		buf.Write(encNil)
		return
	}
	enc.writeSize(len(value), buf)
	for _, item := range value {
		enc.encode(item, true)
	}
}

func (enc *Encoder) writeString(value string, buf *bytes.Buffer) {
	enc.writeSize(len(value), buf)
	buf.Write([]byte(value))
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

func (enc *Encoder) writeStructInfo(value *structInfo, buf *bytes.Buffer) {
	buf.Write(encStructInfo)
	enc.writeUint64(value.id, buf)
	enc.writeSize(value.n, buf)
	for _, f := range value.fields {
		enc.writeString(f.name, buf)
		buf.Write(f.typ)
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
		enc.pad[i] = byte(left)
		i += 1
		if value == 0 {
			break
		}
	}
	buf.Write(enc.pad[:i])
}

func WriteBigDecimal(value *big.Decimal, buf *bytes.Buffer) {
	left, right := value.Components()
	writeBigInt(left, bigintMagicNumber1, buf)
	if right != nil {
		if left.IsNegative() {
			buf.Write([]byte{'\xff'})
		} else {
			buf.Write([]byte{'\x00'})
		}
		writeBigInt(right, bigintMagicNumber2, buf)
	}
}

func WriteBigInt(value *big.Int, buf *bytes.Buffer) {
	writeBigInt(value, bigintMagicNumber1, buf)
}

func writeBigInt(value *big.Int, cutoff *big.Int, buf *bytes.Buffer) {
	if value.IsZero() {
		buf.Write(zeroBase)
		return
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
			buf.Write(encoding)
			return
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
		buf.Write(encoding)
		return
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
		buf.Write(encoding)
		return
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
	buf.Write(encoding)
	return
}

func WriteInt64(value int64) []byte {
	var x uint64
	if value < 0 {
		x = uint64(^value<<1) | 1
	} else {
		x = uint64(value << 1)
	}
	return WriteUint64(x)
}

func WriteInt64AsBig(value int64, buf *bytes.Buffer) {
	if value == 0 {
		buf.Write(zero)
		return
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
			buf.Write(encoding)
			return
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
		buf.Write(encoding)
		return
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
		buf.Write(encoding)
		return
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
	buf.Write(encoding)
	return
}

func WriteSize(value int) ([]byte, os.Error) {
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

func WriteStringNumber(value string, buf *bytes.Buffer) os.Error {
	if strings.Count(value, ".") > 0 {
		number, ok := big.NewDecimal(value)
		if !ok {
			return Error("couldn't create a big.Decimal representation of " + value)
		}
		WriteBigDecimal(number, buf)
		return nil
	}
	number, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return Error("couldn't create an big.Int representation of " + value)
	}
	WriteBigInt(number, buf)
	return nil
}

func WriteUint64(value uint64) []byte {
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

func compileEncEngine(rt reflect.Type) *encEngine {

	if engine, present := encEngines[rt]; present {
		return engine
	}

	engine := &encEngine{}

	switch rt.Kind() {
	case reflect.Array, reflect.Slice:
		engine.elems = make([]*encElem, 1)
		engine.typ = sliceEngine
		elem := engine.setElem(0, rt.Elem())
		elem.typ, _ = getEncType(rt.Elem())
	case reflect.Interface, reflect.Ptr:
		engine.elems = make([]*encElem, 1)
		engine.typ = indirEngine
		engine.setElem(0, rt.Elem())
	case reflect.Struct:
		engine.typ = structEngine
		engine.info = encTypeInfo.Get(rt)
	default:
		error(Error("unknown type to compile an encoding engine: " + rt.String()))
	}

	encEngines[rt] = engine
	return engine

}

func NewEncoder(w io.Writer) *Encoder {
	if b, ok := w.(*bytes.Buffer); ok {
		return &Encoder{
			b:   b,
			dup: false,
			pad: make([]byte, 11),
			w:   w,
		}
	}
	return &Encoder{
		b:   &bytes.Buffer{},
		dup: true,
		pad: make([]byte, 11),
		w:   w,
	}
}

type encEngine struct {
	elems []*encElem
	info  *structInfo
	typ   int
}

type encElem struct {
	engine *encEngine
	op     int
	raw    bool
	typ    []byte
}

type structInfo struct {
	fields []*fieldInfo
	id     uint64
	n      int
}

type fieldInfo struct {
	encElem
	idx    int
	name   string
	nested bool
}

func (engine *encEngine) setElem(idx int, rt reflect.Type) *encElem {
	elem := &encElem{}
	if encRawTypes[rt] {
		elem.raw = true
	} else if op := encBaseOps[rt.Kind()]; op > 0 {
		elem.op = op
	} else {
		elem.engine = compileEncEngine(rt)
	}
	engine.elems[idx] = elem
	return elem
}

type typeRegistry struct {
	nextId uint64
	types  map[reflect.Type]*structInfo
}

func (reg *typeRegistry) Get(t reflect.Type) *structInfo {
	if info, exists := reg.types[t]; exists {
		return info
	}
	s := &structInfo{fields: make([]*fieldInfo, 0), id: reg.nextId}
	n := 0
	reg.nextId += 1
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
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
		f := &fieldInfo{name: name}
		f.typ, f.nested = getEncType(field.Type)
		if encRawTypes[field.Type] {
			f.raw = true
		} else if op := encBaseOps[field.Type.Kind()]; op > 0 {
			f.op = op
		} else {
			f.engine = compileEncEngine(field.Type)
		}
		f.idx = i
		s.fields = append(s.fields, f)
		n += 1
	}
	s.n = n
	return s
}

var encTypeInfo *typeRegistry = &typeRegistry{
	nextId: baseId,
	types:  make(map[reflect.Type]*structInfo),
}

var enctypeMap = [...][]byte{
	reflect.Array:      encAny,
	reflect.Bool:       encBool,
	reflect.Complex64:  encComplex64,
	reflect.Complex128: encComplex128,
	reflect.Float32:    encFloat32,
	reflect.Float64:    encFloat64,
	reflect.Int:        encInt64,
	reflect.Int8:       encByte,
	reflect.Int16:      encInt32,
	reflect.Int32:      encInt32,
	reflect.Int64:      encInt64,
	reflect.Interface:  encAny,
	reflect.Map:        encAny,
	reflect.Ptr:        encAny,
	reflect.String:     encString,
	reflect.Struct:     encAny,
	reflect.Uint:       encUint64,
	reflect.Uint16:     encUint32,
	reflect.Uint32:     encUint32,
	reflect.Uint64:     encUint64,
}

var enctypeNested = [256]bool{
	reflect.Array:     true,
	reflect.Interface: true,
	reflect.Map:       true,
	reflect.Ptr:       true,
	reflect.Struct:    true,
}

func getEncType(t reflect.Type) ([]byte, bool) {
	kind := t.Kind()
	if kind == reflect.Slice {
		if t == byteSliceType {
			return encByteSlice, false
		}
		if t == stringSliceType {
			return encStringSlice, false
		}
		if t == byteSliceSliceType {
			return encByteSliceSlice, false
		}
		return encAny, true
	}
	return enctypeMap[kind], enctypeNested[kind]
}
