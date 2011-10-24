// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	"amp/big"
	"amp/rpc"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	encBigDecimal  = []byte{BigDecimal}
	encBigInt      = []byte{BigInt}
	encByteSlice   = []byte{ByteSlice}
	encDict        = []byte{Dict}
	encFalse       = []byte{False}
	encHeader      = []byte{Header}
	encInt         = []byte{Int}
	encInt64       = []byte{Int64}
	encSlice       = []byte{Slice}
	encString      = []byte{String}
	encStringSlice = []byte{StringSlice}
	encTrue        = []byte{True}
)

type Encoder struct {
	w io.Writer
}

func (enc *Encoder) Encode(v interface{}) (err os.Error) {

	switch value := v.(type) {
	case *big.Decimal:
		_, err = enc.w.Write(encBigDecimal)
		if err != nil {
			return
		}
		return enc.WriteBigDecimal(value)
	case *big.Int:
		_, err = enc.w.Write(encBigInt)
		if err != nil {
			return
		}
		return enc.WriteBigInt(value)
	case bool:
		if value {
			_, err = enc.w.Write(encTrue)
		} else {
			_, err = enc.w.Write(encFalse)
		}
		if err != nil {
			return
		}
		return
	case []byte:
		_, err = enc.w.Write(encByteSlice)
		if err != nil {
			return
		}
		return enc.WriteByteSlice(value)
	case []interface{}:
		_, err = enc.w.Write(encSlice)
		if err != nil {
			return
		}
		return enc.WriteSlice(value)
	case int:
		_, err = enc.w.Write(encInt)
		if err != nil {
			return
		}
		return enc.WriteInt(value)
	case int64:
		_, err = enc.w.Write(encInt64)
		if err != nil {
			return
		}
		return enc.WriteInt64(value)
	case map[string]interface{}:
		_, err = enc.w.Write(encDict)
		if err != nil {
			return
		}
		return enc.WriteDict(value)
	case rpc.Header:
		_, err = enc.w.Write(encHeader)
		if err != nil {
			return
		}
		return enc.WriteDict(value)
	case string:
		_, err = enc.w.Write(encString)
		if err != nil {
			return
		}
		return enc.WriteString(value)
	case []string:
		_, err = enc.w.Write(encStringSlice)
		if err != nil {
			return
		}
		return enc.WriteStringSlice(value)
	}

	msg := fmt.Sprintf("argo: encoding unknown type: %s", v)
	panic(msg)

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
	err = enc.WriteInt(len(value))
	if err != nil {
		return
	}
	_, err = enc.w.Write(value)
	return
}

func (enc *Encoder) WriteDict(value map[string]interface{}) (err os.Error) {
	err = enc.WriteInt(len(value))
	if err != nil {
		return
	}
	for k, v := range value {
		err = enc.WriteInt(len(k))
		if err != nil {
			return
		}
		_, err = enc.w.Write([]byte(k))
		if err != nil {
			return
		}
		err = enc.Encode(v)
		if err != nil {
			return
		}
	}
	return nil
}

func (enc *Encoder) WriteInt(value int) os.Error {
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

func (enc *Encoder) WriteInt64(value int64) os.Error {
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

func (enc *Encoder) WriteSlice(value []interface{}) (err os.Error) {
	err = enc.WriteInt(len(value))
	if err != nil {
		return
	}
	for _, item := range value {
		err = enc.Encode(item)
		if err != nil {
			return
		}
	}
	return
}

func (enc *Encoder) WriteString(value string) (err os.Error) {
	err = enc.WriteInt(len(value))
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
			return Error("Couldn't create a big.Decimal representation of " + value)
		}
		return enc.WriteBigDecimal(number)
	}
	number, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return Error("Couldn't create an big.Int representation of " + value)
	}
	return enc.WriteBigInt(number)
}

func (enc *Encoder) WriteStringSlice(value []string) (err os.Error) {
	err = enc.WriteInt(len(value))
	if err != nil {
		return
	}
	for _, item := range value {
		err = enc.WriteInt(len(item))
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
