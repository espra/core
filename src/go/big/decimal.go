// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package big

import (
	"strings"
)

var (
	nat40    nat
	Decimal0 *Decimal
	Decimal1 *Decimal
)

type Decimal struct {
	a   nat
	b   nat
	neg bool
}

func NewDecimal(value string) (*Decimal, bool) {

	valsize := len(value)
	if valsize == 0 {
		return nil, false
	}

	neg := value[0] == '-'
	if neg || value[0] == '+' {
		value = value[1:valsize]
		valsize -= 1
		if valsize == 0 {
			return nil, false
		}
	}

	// Check if there's a decimal point in the value and exit if there's more
	// than just one decimal point.
	point, set := -1, false
	for i := 0; i < valsize; i++ {
		if value[i] == '.' {
			if set {
				return nil, false
			}
			point = i
			set = true
		}
	}

	var n int
	if set {
		if point != valsize-1 {
			value = value[:point] + value[point+1:]
			n = valsize - point - 1
		} else {
			value = value[:point]
		}
		valsize = len(value)
	}

	a := nat{}.make(0)
	for i := 0; i < valsize; i++ {
		d := hexValue(value[i])
		if 0 <= d && d < 10 {
			a = a.mulAddWW(a, Word(10), Word(d))
		} else {
			return nil, false
		}
	}

	b := nat{}.make(0)
	b = b.mulAddWW(b, Word(10), Word(1))
	if n > 0 {
		for i := 0; i < n; i++ {
			b = b.mulAddWW(b, Word(10), Word(0))
		}
	}

	// Set the negative sign if it exists, except in the case of 0 which has no
	// sign.
	decimal := &Decimal{
		a:   a,
		b:   b,
		neg: len(a) > 0 && neg,
	}

	return decimal, true

}

func DecimalZero() *Decimal {
	return &Decimal{
		a:   nat{}.make(0),
		b:   nat{}.make(0),
		neg: false,
	}
}

func (d *Decimal) String() (s string) {
	if len(d.a) == 0 {
		return "0"
	}
	if d.neg {
		s = "-"
	}
	value := d.a.string10()
	multiplier := len(d.b.string10())
	if multiplier == 1 {
		return s + value
	}
	valsize := len(value)
	diff := multiplier - valsize
	if diff > 0 {
		s += "0"
		rhs := ""
		for i := 0; i < diff-1; i++ {
			rhs += "0"
		}
		rhs = strings.TrimRight(rhs+value, "0")
		if len(rhs) > 0 {
			return s + "." + rhs
		}
		return "0"
	}
	diff = valsize - multiplier + 1
	rhs := strings.TrimRight(value[diff:], "0")
	if len(rhs) > 0 {
		return s + value[:diff] + "." + rhs
	}
	return s + value[:diff]
}

func (d *Decimal) Components() (*Int, *Int) {
	if len(d.a) == 0 {
		return NewIntComponent("0", false), nil
	}
	value := d.a.string10()
	multiplier := len(d.b.string10())
	if multiplier == 1 {
		return NewIntComponent(value, d.neg), nil
	}
	valsize := len(value)
	diff := multiplier - valsize
	if diff > 0 {
		rhs := ""
		for i := 0; i < diff-1; i++ {
			rhs += "0"
		}
		rhs = strings.TrimRight(rhs+value, "0")
		if len(rhs) > 0 {
			return NewIntComponent("0", d.neg), NewIntComponent(pad40("1"+rhs), false)
		}
		return NewIntComponent("0", false), nil
	}
	diff = valsize - multiplier + 1
	rhs := strings.TrimRight(value[diff:], "0")
	if len(rhs) > 0 {
		return NewIntComponent(value[:diff], d.neg), NewIntComponent(pad40("1"+rhs), false)
	}
	return NewIntComponent(value[:diff], false), nil
}

func pad40(s string) string {
	length := len(s)
	if length > 40 {
		s = s[0:40]
	}
	value := make([]byte, 40)
	for i := 0; i < length; i++ {
		value[i] = s[i]
	}
	for i := length; i < 40; i++ {
		value[i] = '0'
	}
	return string(value)
}

func (d *Decimal) Copy() *Decimal {
	a := nat{}
	b := nat{}
	return &Decimal{
		a:   a.set(d.a),
		b:   b.set(d.b),
		neg: d.neg,
	}
}

// Sign returns:
//
//     -1 if x <  0
//      0 if x == 0
//     +1 if x >  0
//
func (d *Decimal) Sign() int {
	if len(d.a) == 0 {
		return 0
	}
	if d.neg {
		return -1
	}
	return 1
}

func (d *Decimal) Abs() *Decimal {
	a := nat{}
	b := nat{}
	return &Decimal{
		a:   a.set(d.a),
		b:   b.set(d.b),
		neg: false,
	}
}

func (d *Decimal) Neg() *Decimal {
	a := nat{}
	b := nat{}
	return &Decimal{
		a:   a.set(d.a),
		b:   b.set(d.b),
		neg: len(d.a) > 0 && !d.neg,
	}
}

func (d *Decimal) IsInt() bool {
	return len(d.b) == 0
}

func (d *Decimal) IsZero() bool {
	return len(d.a) == 0
}

// Cmp compares x and y and returns:
//
//   -1 if x <  y
//    0 if x == y
//   +1 if x >  y
//
func (x *Decimal) Cmp(y *Decimal) (r int) {
	switch {
	case x.neg == y.neg:
		a := nat{}.mul(x.a, y.b)
		b := nat{}.mul(y.a, x.b)
		r = a.cmp(b)
		if x.neg {
			r = -r
		}
	case x.neg:
		r = -1
	default:
		r = 1
	}
	return
}

func (x *Decimal) Add(y *Decimal) *Decimal {
	m := nat{}
	m = m.mul(x.a, y.b)
	neg1 := len(m) > 0 && x.neg
	n := nat{}
	n = n.mul(y.a, x.b)
	neg2 := len(n) > 0 && y.neg
	if neg1 == neg2 {
		m = m.add(m, n)
	} else {
		if m.cmp(n) >= 0 {
			m = m.sub(m, n)
		} else {
			neg1 = !neg1
			m = m.sub(n, m)
		}
	}
	return &Decimal{
		a:   m,
		b:   nat{}.mul(x.b, y.b),
		neg: len(m) > 0 && neg1,
	}
}

func (x *Decimal) Sub(y *Decimal) *Decimal {
	m := nat{}
	m = m.mul(x.a, y.b)
	neg1 := len(m) > 0 && x.neg
	n := nat{}
	n = n.mul(y.a, x.b)
	neg2 := len(n) > 0 && y.neg
	if neg1 != neg2 {
		m = m.add(m, n)
	} else {
		if m.cmp(n) >= 0 {
			m = m.sub(m, n)
		} else {
			neg1 = !neg1
			m = m.sub(n, m)
		}
	}
	return &Decimal{
		a:   m,
		b:   nat{}.mul(x.b, y.b),
		neg: len(m) > 0 && neg1,
	}
}

func (x *Decimal) Mul(y *Decimal) *Decimal {
	a := nat{}.mul(x.a, y.a)
	return &Decimal{
		a:   a,
		b:   nat{}.mul(x.b, y.b),
		neg: len(a) > 0 && x.neg != y.neg,
	}
}

func (x *Decimal) Div(y *Decimal) *Decimal {
	if len(y.a) == 0 {
		panic("division by zero")
	}
	a := nat{}.mul(x.a, y.b)
	b := nat{}.mul(y.a, x.b)
	r := a.cmp(b)
	if r == 0 {
		res := Decimal1.Copy()
		res.neg = x.neg != y.neg
		return res
	}
	a = a.mul(a, nat40)
	if len(b) == 1 {
		a, _ = nat{}.divW(a, b[0])
	} else {
		a, _ = nat{}.divLarge(nat{}, a, b)
	}
	if r == 1 {
		return &Decimal{
			a:   a,
			b:   b.set(nat40),
			neg: x.neg != y.neg,
		}
	}
	return &Decimal{
		a:   a,
		b:   b.set(nat40),
		neg: x.neg != y.neg,
	}
}

func (x nat) string10() string {
	if len(x) == 0 {
		return "0"
	}
	i := x.bitLen()/log2(Word(10)) + 1
	s := make([]byte, i)
	q := nat(nil).set(x)
	for len(q) > 0 {
		i--
		var r Word
		q, r = q.divW(q, Word(10))
		s[i] = "0123456789"[r]
	}
	return string(s[i:])
}

func (x *Int) IsZero() bool {
	return len(x.abs) == 0
}

func (x *Int) RawNeg() bool {
	return x.neg
}

func NewIntString(value string) (*Int, bool) {
	valsize := len(value)
	if valsize == 0 {
		return nil, false
	}
	neg := value[0] == '-'
	if neg || value[0] == '+' {
		value = value[1:]
		valsize = len(value)
		if valsize == 0 {
			return nil, false
		}
	}
	a := nat{}.make(0)
	for i := 0; i < valsize; i++ {
		d := hexValue(value[i])
		if 0 <= d && d < 10 {
			a = a.mulAddWW(a, Word(10), Word(d))
		} else {
			return nil, false
		}
	}
	return &Int{
		abs: a,
		neg: len(a) > 0 && neg,
	},
		true
}

func NewIntComponent(value string, neg bool) *Int {
	valsize := len(value)
	a := nat{}.make(0)
	for i := 0; i < valsize; i++ {
		d := hexValue(value[i])
		if 0 <= d && d < 10 {
			a = a.mulAddWW(a, Word(10), Word(d))
		}
	}
	return &Int{
		abs: a,
		neg: neg,
	}
}

func init() {
	nat40 = nat{}.make(0)
	nat40 = nat40.mulAddWW(nat40, Word(10), Word(1))
	for i := 0; i < 40; i++ {
		nat40 = nat40.mulAddWW(nat40, Word(10), Word(0))
	}
	Decimal0, _ = NewDecimal("0")
	Decimal1, _ = NewDecimal("1")
}
