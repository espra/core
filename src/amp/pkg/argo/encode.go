// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	"amp/big"
	"amp/rpc"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"strings"
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
	encUint32      = []byte{Uint32}
	encUint64      = []byte{Uint64}
)

var (
	anyInterfaceType = reflect.TypeOf(interface{}(nil))
	byteSliceType    = reflect.TypeOf([]byte(nil))
	stringSliceType  = reflect.TypeOf([]string(nil))
)

type Encoder struct {
	w io.Writer
}

func (enc *Encoder) Encode(v interface{}) (err os.Error) {
	return enc.encode(v, true)
}

func (enc *Encoder) encode(v interface{}, typeinfo bool) (err os.Error) {

	switch value := v.(type) {
	case *big.Decimal:
		if typeinfo {
			_, err = enc.w.Write(encBigDecimal)
			if err != nil {
				return
			}
		}
		return enc.WriteBigDecimal(value)
	case *big.Int:
		if typeinfo {
			_, err = enc.w.Write(encBigInt)
			if err != nil {
				return
			}
		}
		return enc.WriteBigInt(value)
	case bool:
		if value {
			_, err = enc.w.Write(encBoolTrue)
		} else {
			_, err = enc.w.Write(encBoolFalse)
		}
		return
	case byte:
		if typeinfo {
			_, err = enc.w.Write(encByte)
			if err != nil {
				return
			}
		}
		_, err = enc.w.Write([]byte{value})
		return
	case []byte:
		if typeinfo {
			_, err = enc.w.Write(encByteSlice)
			if err != nil {
				return
			}
		}
		return enc.WriteByteSlice(value)
	case complex64:
		if typeinfo {
			_, err = enc.w.Write(encComplex64)
			if err != nil {
				return
			}
		}
		err = enc.WriteFloat32(real(value))
		if err != nil {
			return err
		}
		return enc.WriteFloat32(imag(value))
	case complex128:
		if typeinfo {
			_, err = enc.w.Write(encComplex128)
			if err != nil {
				return
			}
		}
		err = enc.WriteFloat64(real(value))
		if err != nil {
			return err
		}
		return enc.WriteFloat64(imag(value))
	case float32:
		if typeinfo {
			_, err = enc.w.Write(encFloat32)
			if err != nil {
				return
			}
		}
		return enc.WriteFloat32(value)
	case float64:
		if typeinfo {
			_, err = enc.w.Write(encFloat32)
			if err != nil {
				return
			}
		}
		return enc.WriteFloat64(value)
	case int:
		if typeinfo {
			_, err = enc.w.Write(encInt64)
			if err != nil {
				return
			}
		}
		return enc.WriteInt64(int64(value))
	case int16:
		if typeinfo {
			_, err = enc.w.Write(encInt32)
			if err != nil {
				return
			}
		}
		return enc.WriteInt64(int64(value))
	case int32:
		if typeinfo {
			_, err = enc.w.Write(encInt32)
			if err != nil {
				return
			}
		}
		return enc.WriteInt64(int64(value))
	case int64:
		if typeinfo {
			_, err = enc.w.Write(encInt64)
			if err != nil {
				return
			}
		}
		return enc.WriteInt64(value)
	case []interface{}:
		if typeinfo {
			_, err = enc.w.Write(encSliceAny)
			if err != nil {
				return
			}
		}
		return enc.WriteSlice(value)
	case map[string]interface{}:
		if typeinfo {
			_, err = enc.w.Write(encDictAny)
			if err != nil {
				return
			}
		}
		return enc.WriteDict(value)
	case rpc.Header:
		if typeinfo {
			_, err = enc.w.Write(encHeader)
			if err != nil {
				return
			}
		}
		return enc.WriteDict(value)
	case string:
		if typeinfo {
			_, err = enc.w.Write(encString)
			if err != nil {
				return
			}
		}
		return enc.WriteString(value)
	case []string:
		if typeinfo {
			_, err = enc.w.Write(encStringSlice)
			if err != nil {
				return
			}
		}
		return enc.WriteStringSlice(value)
	case uint16:
		if typeinfo {
			_, err = enc.w.Write(encUint32)
			if err != nil {
				return
			}
		}
		return enc.WriteUint64(uint64(value))
	case uint32:
		if typeinfo {
			_, err = enc.w.Write(encUint32)
			if err != nil {
				return
			}
		}
		return enc.WriteUint64(uint64(value))
	case uint64:
		if typeinfo {
			_, err = enc.w.Write(encUint64)
			if err != nil {
				return
			}
		}
		return enc.WriteUint64(value)
	}

	return enc.encodeValue(reflect.ValueOf(v), typeinfo)

}

func (enc *Encoder) encodeValue(v reflect.Value, typeinfo bool) (err os.Error) {

	if !v.IsValid() {
		return Error("invalid type detected")
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		typ := v.Type()
		if typ == byteSliceType {
			if typeinfo {
				_, err = enc.w.Write(encByteSlice)
				if err != nil {
					return
				}
			}
			return enc.WriteByteSlice(v.Interface().([]byte))
		}
		if typ == stringSliceType {
			if typeinfo {
				_, err = enc.w.Write(encStringSlice)
				if err != nil {
					return
				}
			}
			return enc.WriteStringSlice(v.Interface().([]string))
		}
		_, err = enc.w.Write(encSlice)
		if err != nil {
			return
		}
		nested, _err := enc.writeEncType(typ.Elem())
		if _err != nil {
			return _err
		}
		if v.IsNil() {
			_, err = enc.w.Write(encNil)
			return
		}
		size := v.Len()
		err = enc.WriteSize(size)
		if err != nil {
			return
		}
		for i := 0; i < size; i++ {
			err = enc.encodeValue(v.Index(i), nested)
			if err != nil {
				return err
			}
		}
		return
	case reflect.Bool:
		if v.Bool() {
			_, err = enc.w.Write(encBoolTrue)
		} else {
			_, err = enc.w.Write(encBoolFalse)
		}
		return
	case reflect.Complex64:
		if typeinfo {
			_, err = enc.w.Write(encComplex64)
			if err != nil {
				return
			}
		}
		val := v.Complex()
		err = enc.WriteFloat32(float32(real(val)))
		if err != nil {
			return err
		}
		return enc.WriteFloat32(float32(imag(val)))
	case reflect.Complex128:
		if typeinfo {
			_, err = enc.w.Write(encComplex128)
			if err != nil {
				return
			}
		}
		val := v.Complex()
		err = enc.WriteFloat64(real(val))
		if err != nil {
			return err
		}
		return enc.WriteFloat64(imag(val))
	case reflect.Float32:
		if typeinfo {
			_, err = enc.w.Write(encFloat32)
			if err != nil {
				return
			}
		}
		return enc.WriteFloat32(float32(v.Float()))
	case reflect.Float64:
		if typeinfo {
			_, err = enc.w.Write(encFloat64)
			if err != nil {
				return
			}
		}
		return enc.WriteFloat64(v.Float())
	case reflect.Int, reflect.Int64:
		val := v.Int()
		if typeinfo {
			_, err = enc.w.Write(encInt64)
			if err != nil {
				return
			}
		}
		return enc.WriteInt64(int64(val))
	case reflect.Int16, reflect.Int32:
		val := v.Int()
		if typeinfo {
			_, err = enc.w.Write(encInt32)
			if err != nil {
				return
			}
		}
		return enc.WriteInt64(int64(val))
	case reflect.Int8:
		if typeinfo {
			_, err = enc.w.Write(encByte)
			if err != nil {
				return
			}
		}
		_, err = enc.w.Write([]byte{byte(v.Int())})
		return
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			_, err = enc.w.Write(encNil)
			return
		}
		return enc.encodeValue(v.Elem(), typeinfo)
	case reflect.Map:
		var forcekey bool
		typ := v.Type()
		keytype := typ.Key()
		elemtype := typ.Elem()
		if keytype.Kind() == reflect.String {
			if elemtype == anyInterfaceType {
				_, err = enc.w.Write(encDictAny)
				if err != nil {
					return
				}
				return enc.WriteDict(v.Interface().(map[string]interface{}))
			} else {
				_, err = enc.w.Write(encDict)
				if err != nil {
					return
				}
			}
		} else {
			_, err = enc.w.Write(encMap)
			if err != nil {
				return
			}
			if keytype == anyInterfaceType {
				_, err = enc.w.Write(encAny)
				if err != nil {
					return
				}
				forcekey = true
			} else {
				forcekey, err = enc.writeEncType(keytype)
				if err != nil {
					return
				}
			}
		}
		var forceval bool
		if elemtype == anyInterfaceType {
			_, err = enc.w.Write(encAny)
			if err != nil {
				return
			}
			forceval = true
		} else {
			forceval, err = enc.writeEncType(elemtype)
			if err != nil {
				return
			}
		}
		if v.IsNil() {
			_, err = enc.w.Write(encNil)
			return
		}
		err = enc.WriteSize(v.Len())
		if err != nil {
			return
		}
		for _, key := range v.MapKeys() {
			err = enc.encodeValue(key, forcekey)
			if err != nil {
				return
			}
			err = enc.encodeValue(v.MapIndex(key), forceval)
			if err != nil {
				return
			}
		}
		return
	case reflect.String:
		_, err = enc.w.Write(encString)
		if err != nil {
			return
		}
		return enc.WriteString(v.String())
	case reflect.Uint, reflect.Uint64:
		val := v.Uint()
		if typeinfo {
			_, err = enc.w.Write(encInt64)
			if err != nil {
				return
			}
		}
		return enc.WriteUint64(val)
	case reflect.Uint16, reflect.Uint32:
		val := v.Uint()
		if typeinfo {
			_, err = enc.w.Write(encInt32)
			if err != nil {
				return
			}
		}
		return enc.WriteUint64(val)
	}

	msg := fmt.Sprintf("argo: encoding unknown type: %s", v.Kind().String())
	panic(msg)

}

func (enc *Encoder) writeEncType(t reflect.Type) (nested bool, err os.Error) {

	var enctype []byte

	switch t.Kind() {
	case reflect.Array, reflect.Slice:
		if t == byteSliceType {
			enctype = encByteSlice
		} else if t == stringSliceType {
			enctype = encStringSlice
		} else {
			enctype = encAny
			nested = true
		}
	case reflect.Bool:
		enctype = encBool
	case reflect.Complex64:
		enctype = encComplex64
	case reflect.Complex128:
		enctype = encComplex128
	case reflect.Float32:
		enctype = encFloat32
	case reflect.Float64:
		enctype = encFloat64
	case reflect.Int, reflect.Int64:
		enctype = encInt64
	case reflect.Int16, reflect.Int32:
		enctype = encInt32
	case reflect.Int8:
		enctype = encByte
	case reflect.Interface, reflect.Ptr:
		enctype = encAny
		nested = true
	case reflect.Map:
		enctype = encAny
		nested = true
	case reflect.String:
		enctype = encString
	case reflect.Uint, reflect.Uint64:
		enctype = encUint64
	case reflect.Uint16, reflect.Uint32:
		enctype = encUint32
	}

	if enctype == nil {
		err = Error("unknown element type: " + t.String())
		return
	}
	_, err = enc.w.Write(enctype)
	return

}

func (enc *Encoder) WriteBigDecimal(value *big.Decimal) (err os.Error) {
	left, right := value.Components()
	err = enc.writeBigInt(left, bigintMagicNumber1)
	if err != nil {
		return
	}
	if right != nil {
		if left.IsNegative() {
			_, err = enc.w.Write([]byte{'\xff'})
		} else {
			_, err = enc.w.Write([]byte{'\x00'})
		}
		if err != nil {
			return
		}
		return enc.writeBigInt(right, bigintMagicNumber2)
	}
	return
}

func (enc *Encoder) WriteBigInt(value *big.Int) (err os.Error) {
	return enc.writeBigInt(value, bigintMagicNumber1)
}

func (enc *Encoder) writeBigInt(value *big.Int, cutoff *big.Int) (err os.Error) {
	if value.IsZero() {
		_, err = enc.w.Write(zeroBase)
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
			_, err = enc.w.Write(encoding)
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
		_, err = enc.w.Write(encoding)
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
		_, err = enc.w.Write(encoding)
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
	_, err = enc.w.Write(encoding)
	return
}

func (enc *Encoder) WriteByteSlice(value []byte) (err os.Error) {
	if value == nil {
		_, err = enc.w.Write(encNil)
		return
	}
	err = enc.WriteSize(len(value))
	if err != nil {
		return
	}
	_, err = enc.w.Write(value)
	return
}

func (enc *Encoder) WriteDict(value map[string]interface{}) (err os.Error) {
	if value == nil {
		_, err = enc.w.Write(encNil)
		return
	}
	err = enc.WriteSize(len(value))
	if err != nil {
		return
	}
	for k, v := range value {
		err = enc.WriteSize(len(k))
		if err != nil {
			return
		}
		_, err = enc.w.Write([]byte(k))
		if err != nil {
			return
		}
		err = enc.encode(v, true)
		if err != nil {
			return
		}
	}
	return nil
}

func (enc *Encoder) WriteFloat32(value float32) (err os.Error) {
	v := math.Float32bits(value)
	data := make([]byte, 4)
	data[0] = byte(v >> 24)
	data[1] = byte(v >> 16)
	data[2] = byte(v >> 8)
	data[3] = byte(v)
	_, err = enc.w.Write(data)
	return
}

func (enc *Encoder) WriteFloat64(value float64) (err os.Error) {
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
	_, err = enc.w.Write(data)
	return
}

func (enc *Encoder) WriteInt64(value int64) os.Error {
	var x uint64
	if value < 0 {
		x = uint64(^value<<1) | 1
	} else {
		x = uint64(value << 1)
	}
	return enc.WriteUint64(x)
}

func (enc *Encoder) WriteInt64AsBig(value int64) (err os.Error) {
	if value == 0 {
		_, err = enc.w.Write(zero)
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
			_, err = enc.w.Write(encoding)
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
		_, err = enc.w.Write(encoding)
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
		_, err = enc.w.Write(encoding)
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
	_, err = enc.w.Write(encoding)
	return
}

func (enc *Encoder) WriteSize(value int) os.Error {
	if value < 0 || value > maxInt32 {
		return OutOfRangeError
	}
	data := make([]byte, 0)
	for {
		leftBits := value & 127
		value >>= 7
		if value > 0 {
			leftBits += 128
		}
		data = append(data, byte(leftBits))
		if value == 0 {
			break
		}
	}
	_, err := enc.w.Write(data)
	return err
}

func (enc *Encoder) WriteSlice(value []interface{}) (err os.Error) {
	if value == nil {
		_, err = enc.w.Write(encNil)
		return
	}
	err = enc.WriteSize(len(value))
	if err != nil {
		return
	}
	for _, item := range value {
		err = enc.encode(item, true)
		if err != nil {
			return
		}
	}
	return
}

func (enc *Encoder) WriteString(value string) (err os.Error) {
	err = enc.WriteSize(len(value))
	if err != nil {
		return
	}
	_, err = enc.w.Write([]byte(value))
	return
}

func (enc *Encoder) WriteStringNumber(value string) (err os.Error) {
	if strings.Count(value, ".") > 0 {
		number, ok := big.NewDecimal(value)
		if !ok {
			return Error("couldn't create a big.Decimal representation of " + value)
		}
		return enc.WriteBigDecimal(number)
	}
	number, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return Error("couldn't create an big.Int representation of " + value)
	}
	return enc.WriteBigInt(number)
}

func (enc *Encoder) WriteStringSlice(value []string) (err os.Error) {
	if value == nil {
		_, err = enc.w.Write(encNil)
		return
	}
	err = enc.WriteSize(len(value))
	if err != nil {
		return
	}
	for _, item := range value {
		err = enc.WriteSize(len(item))
		if err != nil {
			return
		}
		_, err = enc.w.Write([]byte(item))
		if err != nil {
			return
		}
	}
	return nil
}

func (enc *Encoder) WriteUint64(value uint64) os.Error {
	data := make([]byte, 0)
	for {
		leftBits := value & 127
		value >>= 7
		if value > 0 {
			leftBits += 128
		}
		data = append(data, byte(leftBits))
		if value == 0 {
			break
		}
	}
	_, err := enc.w.Write(data)
	return err
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
	return &Encoder{w: w}
}
