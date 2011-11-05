// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	"amp/big"
	"amp/rpc"
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"utf8"
)

const maxInt32 = 2147483647

var (
	encAny         = []byte{Any}
	encBigDecimal  = []byte{BigDecimal}
	encBigInt      = []byte{BigInt}
	encBool        = []byte{Bool}
	encBoolFalse   = []byte{BoolFalse}
	encBoolTrue    = []byte{BoolTrue}
	encByte        = []byte{Byte}
	encByteSlice   = []byte{ByteSlice}
	encComplex64   = []byte{Complex64}
	encComplex128  = []byte{Complex128}
	encDict        = []byte{Dict}
	encDictAny     = []byte{Dict, Any}
	encFloat32     = []byte{Float32}
	encFloat64     = []byte{Float64}
	encHeader      = []byte{Header}
	encInt32       = []byte{Int32}
	encInt64       = []byte{Int64}
	encItem        = []byte{Item}
	encMap         = []byte{Map}
	encNil         = []byte{Nil}
	encSlice       = []byte{Slice}
	encSliceAny    = []byte{Slice, Any}
	encString      = []byte{String}
	encStringSlice = []byte{StringSlice}
	encStruct      = []byte{Struct}
	encStructInfo  = []byte{StructInfo}
	encUint32      = []byte{Uint32}
	encUint64      = []byte{Uint64}
)

var (
	anyInterfaceType = reflect.TypeOf(interface{}(nil))
	byteSliceType    = reflect.TypeOf([]byte(nil))
	stringSliceType  = reflect.TypeOf([]string(nil))
)

type Encoder struct {
	b    *bytes.Buffer
	dup  bool
	err  os.Error
	w    io.Writer
	seen map[reflect.Type]*structInfo
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
		WriteByteSlice(value, enc.b)
	case complex64:
		if typeinfo {
			enc.b.Write(encComplex64)
		}
		WriteFloat32(real(value), enc.b)
		WriteFloat32(imag(value), enc.b)
	case complex128:
		if typeinfo {
			enc.b.Write(encComplex128)
		}
		WriteFloat64(real(value), enc.b)
		WriteFloat64(imag(value), enc.b)
	case float32:
		if typeinfo {
			enc.b.Write(encFloat32)
		}
		WriteFloat32(value, enc.b)
	case float64:
		if typeinfo {
			enc.b.Write(encFloat64)
		}
		WriteFloat64(value, enc.b)
	case int:
		if typeinfo {
			enc.b.Write(encInt64)
		}
		WriteInt64(int64(value), enc.b)
	case int16:
		if typeinfo {
			enc.b.Write(encInt32)
		}
		WriteInt64(int64(value), enc.b)
	case int32:
		if typeinfo {
			enc.b.Write(encInt32)
		}
		WriteInt64(int64(value), enc.b)
	case int64:
		if typeinfo {
			enc.b.Write(encInt64)
		}
		WriteInt64(value, enc.b)
	case []interface{}:
		if typeinfo {
			enc.b.Write(encSliceAny)
		}
		enc.WriteSlice(value, enc.b)
	case map[string]interface{}:
		if typeinfo {
			enc.b.Write(encDictAny)
		}
		enc.WriteDict(value, enc.b)
	case rpc.Header:
		if typeinfo {
			enc.b.Write(encHeader)
		}
		enc.WriteDict(value, enc.b)
	case string:
		if typeinfo {
			enc.b.Write(encString)
		}
		WriteString(value, enc.b)
	case []string:
		if typeinfo {
			enc.b.Write(encStringSlice)
		}
		WriteStringSlice(value, enc.b)
	case *structInfo:
		enc.b.Write(encStructInfo)
		WriteUint64(value.id, enc.b)
		WriteSize(value.n, enc.b)
		for _, f := range value.fields {
			if f == nil {
				continue
			}
			WriteString(f.name, enc.b)
			enc.b.Write(f.enctype)
		}
	case uint16:
		if typeinfo {
			enc.b.Write(encUint32)
		}
		WriteUint64(uint64(value), enc.b)
	case uint32:
		if typeinfo {
			enc.b.Write(encUint32)
		}
		WriteUint64(uint64(value), enc.b)
	case uint64:
		if typeinfo {
			enc.b.Write(encUint64)
		}
		WriteUint64(value, enc.b)
	default:
		enc.encodeValue(reflect.ValueOf(v), typeinfo)
	}

}

func (enc *Encoder) encodeValue(v reflect.Value, typeinfo bool) {

	if !v.IsValid() {
		enc.err = Error("invalid type detected")
		return
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		typ := v.Type()
		if typ == byteSliceType {
			if typeinfo {
				enc.b.Write(encByteSlice)
			}
			WriteByteSlice(v.Interface().([]byte), enc.b)
		}
		if typ == stringSliceType {
			if typeinfo {
				enc.b.Write(encStringSlice)
			}
			WriteStringSlice(v.Interface().([]string), enc.b)
		}
		enc.b.Write(encSlice)
		nested := enc.writeEncType(typ.Elem())
		if v.IsNil() {
			enc.b.Write(encNil)
			return
		}
		size := v.Len()
		WriteSize(size, enc.b)
		for i := 0; i < size; i++ {
			enc.encodeValue(v.Index(i), nested)
		}
	case reflect.Bool:
		if v.Bool() {
			enc.b.Write(encBoolTrue)
		} else {
			enc.b.Write(encBoolFalse)
		}
	case reflect.Complex64:
		if typeinfo {
			enc.b.Write(encComplex64)
		}
		val := v.Complex()
		WriteFloat32(float32(real(val)), enc.b)
		WriteFloat32(float32(imag(val)), enc.b)
	case reflect.Complex128:
		if typeinfo {
			enc.b.Write(encComplex128)
		}
		val := v.Complex()
		WriteFloat64(real(val), enc.b)
		WriteFloat64(imag(val), enc.b)
	case reflect.Float32:
		if typeinfo {
			enc.b.Write(encFloat32)
		}
		WriteFloat32(float32(v.Float()), enc.b)
	case reflect.Float64:
		if typeinfo {
			enc.b.Write(encFloat64)
		}
		WriteFloat64(v.Float(), enc.b)
	case reflect.Int, reflect.Int64:
		if typeinfo {
			enc.b.Write(encInt64)
		}
		WriteInt64(v.Int(), enc.b)
	case reflect.Int16, reflect.Int32:
		if typeinfo {
			enc.b.Write(encInt32)
		}
		WriteInt64(v.Int(), enc.b)
	case reflect.Int8:
		if typeinfo {
			enc.b.Write(encByte)
		}
		enc.b.Write([]byte{byte(v.Int())})
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			enc.b.Write(encNil)
			return
		}
		enc.encodeValue(v.Elem(), typeinfo)
	case reflect.Map:
		var forcekey bool
		typ := v.Type()
		keytype := typ.Key()
		elemtype := typ.Elem()
		if keytype.Kind() == reflect.String {
			if elemtype == anyInterfaceType {
				enc.b.Write(encDictAny)
				enc.WriteDict(v.Interface().(map[string]interface{}), enc.b)
				return
			} else {
				enc.b.Write(encDict)
			}
		} else {
			enc.b.Write(encMap)
			if keytype == anyInterfaceType {
				enc.b.Write(encAny)
				forcekey = true
			} else {
				forcekey = enc.writeEncType(keytype)
			}
		}
		var forceval bool
		if elemtype == anyInterfaceType {
			enc.b.Write(encAny)
			forceval = true
		} else {
			forceval = enc.writeEncType(elemtype)
		}
		if v.IsNil() {
			enc.b.Write(encNil)
			return
		}
		WriteSize(v.Len(), enc.b)
		for _, key := range v.MapKeys() {
			enc.encodeValue(key, forcekey)
			enc.encodeValue(v.MapIndex(key), forceval)
		}
	case reflect.String:
		enc.b.Write(encString)
		WriteString(v.String(), enc.b)
	case reflect.Struct:
		typ := v.Type()
		if enc.seen == nil {
			enc.seen = make(map[reflect.Type]*structInfo)
		}
		info, ok := enc.seen[typ]
		if !ok {
			info = typeCache.Get(typ)
			enc.encode(info, true)
			enc.seen[typ] = info
		}
		enc.b.Write(encStruct)
		WriteUint64(info.id, enc.b)
		for i, f := range info.fields {
			if f == nil {
				continue
			}
			enc.encodeValue(v.Field(i), f.nested)
		}
	case reflect.Uint, reflect.Uint64:
		if typeinfo {
			enc.b.Write(encInt64)
		}
		WriteUint64(v.Uint(), enc.b)
	case reflect.Uint16, reflect.Uint32:
		if typeinfo {
			enc.b.Write(encInt32)
		}
		WriteUint64(v.Uint(), enc.b)
	default:
		msg := fmt.Sprintf("argo: encoding unknown type: %s", v.Kind().String())
		panic(msg)
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

func WriteByteSlice(value []byte, buf *bytes.Buffer) {
	if value == nil {
		buf.Write(encNil)
		return
	}
	WriteSize(len(value), buf)
	buf.Write(value)
}

func (enc *Encoder) WriteDict(value map[string]interface{}, buf *bytes.Buffer) {
	if value == nil {
		buf.Write(encNil)
		return
	}
	WriteSize(len(value), buf)
	for k, v := range value {
		WriteSize(len(k), buf)
		buf.Write([]byte(k))
		enc.encode(v, true)
	}
}

func WriteFloat32(value float32, buf *bytes.Buffer) {
	v := math.Float32bits(value)
	data := make([]byte, 4)
	data[0] = byte(v >> 24)
	data[1] = byte(v >> 16)
	data[2] = byte(v >> 8)
	data[3] = byte(v)
	buf.Write(data)
}

func WriteFloat64(value float64, buf *bytes.Buffer) {
	v := math.Float64bits(value)
	data := make([]byte, 8)
	data[0] = byte(v >> 56)
	data[1] = byte(v >> 48)
	data[2] = byte(v >> 40)
	data[3] = byte(v >> 32)
	data[4] = byte(v >> 24)
	data[5] = byte(v >> 16)
	data[6] = byte(v >> 8)
	data[7] = byte(v)
	buf.Write(data)
}

func WriteInt64(value int64, buf *bytes.Buffer) {
	var x uint64
	if value < 0 {
		x = uint64(^value<<1) | 1
	} else {
		x = uint64(value << 1)
	}
	WriteUint64(x, buf)
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

func WriteSize(value int, buf *bytes.Buffer) {
	if value < 0 || value > maxInt32 {
		error(OutOfRangeError)
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
	buf.Write(data[:i])
}

func (enc *Encoder) WriteSlice(value []interface{}, buf *bytes.Buffer) {
	if value == nil {
		buf.Write(encNil)
		return
	}
	WriteSize(len(value), buf)
	for _, item := range value {
		enc.encode(item, true)
	}
}

func WriteString(value string, buf *bytes.Buffer) {
	WriteSize(len(value), buf)
	buf.Write([]byte(value))
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

func WriteStringSlice(value []string, buf *bytes.Buffer) {
	if value == nil {
		buf.Write(encNil)
		return
	}
	WriteSize(len(value), buf)
	for _, item := range value {
		WriteSize(len(item), buf)
		buf.Write([]byte(item))
	}
}

func WriteUint64(value uint64, buf *bytes.Buffer) {
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
	buf.Write(data[:i])
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
		return &Encoder{w: w, b: b, dup: false}
	}
	return &Encoder{w: w, b: &bytes.Buffer{}, dup: true}
}

type structInfo struct {
	fields []*fieldInfo
	id     uint64
	n      int
}

type fieldInfo struct {
	enctype []byte
	name    string
	nested  bool
}

type typeRegistry struct {
	nextId uint64
	lock   sync.Mutex
	types  map[reflect.Type]*structInfo
}

func (reg *typeRegistry) Get(t reflect.Type) *structInfo {
	reg.lock.Lock()
	defer reg.lock.Unlock()
	if info, exists := reg.types[t]; exists {
		return info
	}
	n := t.NumField()
	s := &structInfo{fields: make([]*fieldInfo, n), id: reg.nextId}
	j := 0
	reg.nextId += 1
	for i := 0; i < n; i++ {
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
		f.enctype, f.nested = getEncType(field.Type)
		s.fields[i] = f
		j += 1
	}
	s.n = j
	return s
}

var typeCache *typeRegistry = &typeRegistry{
	types: make(map[reflect.Type]*structInfo),
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
		return encAny, true
	}
	return enctypeMap[kind], enctypeNested[kind]
}

func error(err os.Error) {
	panic(err.String())
}
