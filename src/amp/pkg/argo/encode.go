// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	"amp/big"
	"io"
	"os"
	"strings"
)

type Encoder struct {
	w io.Writer
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

func (enc *Encoder) WriteString(value string) (err os.Error) {
	err = enc.WriteInt(len(value))
	if err != nil {
		return
	}
	_, err = enc.w.Write([]byte(value))
	return
}

func (enc *Encoder) WriteStringList(value []string) (err os.Error) {
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

func (enc *Encoder) WriteInt64(value int64) (err os.Error) {
	if value == 0 {
		_, err = enc.w.Write(zero)
		return
	}
	if value > 0 {
		if value < magicNumber {
			encoding := []byte{'\x01', '\x80', '\x01', '\x01'}
			div, mod := value/255, value%255
			encoding[3] = byte(mod) + 1
			if div > 0 {
				div, mod = div/255, div%255
				encoding[2] = byte(mod) + 1
				if div > 0 {
					encoding[1] = byte(div) + 128
				}
			}
			_, err = enc.w.Write(encoding)
			return
		}
		value -= magicNumber
		encoding := []byte{'\x01', '\xff'}
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
		encoding := []byte{'\x01', '\x7f', '\xfe', '\xfe'}
		div, mod := value/255, value%255
		encoding[3] = 254 - byte(mod)
		if div > 0 {
			div, mod = div/255, div%255
			encoding[2] = 254 - byte(mod)
			if div > 0 {
				encoding[1] = 127 - byte(div)
			}
		}
		_, err = enc.w.Write(encoding)
		return
	}
	value -= magicNumber
	encoding := []byte{'\x01', '\x00'}
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

func (enc *Encoder) WriteBigInt(value *big.Int) (err os.Error) {
	_, err = enc.w.Write([]byte{'\x01'})
	if err != nil {
		return
	}
	return enc.writeBigInt(value, bigintMagicNumber1)
}

func (enc *Encoder) WriteBigDecimal(value *big.Decimal) (err os.Error) {
	_, err = enc.w.Write([]byte{'\x01'})
	if err != nil {
		return
	}
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

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}
