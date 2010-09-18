// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package big

import (
	"strings"
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
			value = value[0:point] + value[point+1:]
			n = valsize - point - 1
		} else {
			value = value[0:point]
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
	if n > 0 {
		b = b.mulAddWW(b, Word(10), Word(1))
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
	if d.neg {
		s = "-"
	}
	if len(d.b) > 0 {
		value := d.a.string(10)
		valsize := len(value)
		multiplier := len(d.b.string(10))
		if multiplier == valsize {
			return s + "0." + strings.TrimRight(value, "0")
		}
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
			return s
		}
		diff = valsize - multiplier
		rhs := strings.TrimRight(value[diff:], "0")
		if len(rhs) > 0 {
			return s + value[0:diff] + "." + rhs
		} else {
			return s + value[0:diff]
		}
	}
	return s + d.a.string(10)
}

func (d *Decimal) Copy() *Decimal {
	a := nat{}.make(0)
	b := nat{}.make(0)
	a.set(d.a)
	b.set(d.b)
	return &Decimal{
		a:   a,
		b:   b,
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

func (d *Decimal) Abs() {
	d.neg = false
}

func (d *Decimal) Neg() {
	d.neg = len(d.a) > 0 && !d.neg
}

func (d *Decimal) IsInt() bool {
	return len(d.b) == 0
}

func (d *Decimal) Add(x *Decimal, y *Decimal) *Decimal {
	x1 := nat{}.make(0)
	x1 = x1.mul(x.a, y.b)
	neg1 := len(x1) > 0 && x.neg
	y1 := nat{}.make(0)
	y1 = y1.mul(y.a, x.b)
	neg2 := len(y1) > 0 && y.neg
	if neg1 == neg2 {
		d.a = d.a.add(x1, y1)
	} else {
		if x1.cmp(y1) >= 0 {
			d.a = d.a.sub(x1, y1)
		} else {
			neg1 = !neg1
			d.a = d.a.sub(y1, x1)
		}
	}
	d.b = d.b.mul(x.b, y.b)
	d.neg = len(d.a) > 0 && neg1
	return d
}

func (d *Decimal) Sub(x *Decimal, y *Decimal) *Decimal {
	x1 := nat{}.make(0)
	x1 = x1.mul(x.a, y.b)
	neg1 := len(x1) > 0 && x.neg
	y1 := nat{}.make(0)
	y1 = y1.mul(y.a, x.b)
	neg2 := len(y1) > 0 && y.neg
	if neg1 != neg2 {
		d.a = d.a.add(x1, y1)
	} else {
		if x1.cmp(y1) >= 0 {
			d.a = d.a.sub(x1, y1)
		} else {
			neg1 = !neg1
			d.a = d.a.sub(y1, x1)
		}
	}
	d.b = d.b.mul(x.b, y.b)
	d.neg = len(d.a) > 0 && neg1
	return d
}
