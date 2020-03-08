// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Package big implements arbitrary-precision decimals.
//
// Unlike other decimal packages, we opt for a cleaner API for developers:
//
// * Instead of returning NaNs on invalid operations, methods like Div and Sqrt
//   return explicit errors instead. This makes it more obvious to developers
//   when they should check for error conditions.
//
// * Instead of rounding on every operation, we only round on specific methods
//   like Div, Sqrt, and String, but otherwise do lossless calculations. If a
//   developer desires rounding after every operation, they can call the
//   explicit Round method.
//
// * Instead of having a max precision set in a configurable context that has
//   to be constantly passed around, all operations default to infinite
//   precision. If explicit precision is desired, then the ChangePrec method
//   can be used with an explicit rounding mode after every operation.
//
// * Similarly, instead of having a max scale and rounding method set in a
//   context, methods like Div, Sqrt, and String default to a max scale of 20,
//   and ToNearestEven rounding. If explicit max scale and rounding mode are
//   desired, then there the corresponding DivExplicit, SqrtExplicit, and
//   StringExplicit methods can be used instead.
package big

import (
	"errors"
	"fmt"
	"math/big"
)

// DefaultMaxScale defines the default maximum scale, i.e. the number of digits
// to the right of the decimal point in a number.
const DefaultMaxScale = 20

// Rounding modes.
const (
	AwayFromZero Rounding = iota
	ToNearestEven
	ToNegativeInf
	ToPositiveInf
	ToZero
)

// Error values.
var (
	ErrDivisionByZero = errors.New("big: division by zero")
	ErrSqrtOfNegative = errors.New("big: sqrt of negative number")
)

// Decimal represents an arbitrary-precision decimal.
type Decimal struct {
	base   uint64
	excess big.Int
	state  uint64
}

// Abs sets d to the absolute value of x, and returns d.
func (d *Decimal) Abs(x *Decimal) *Decimal {
	return d
}

// Add sets d to the sum of x and y, and returns d.
func (d *Decimal) Add(x *Decimal, y *Decimal) *Decimal {
	return d
}

// ChangePrec changes d's precision to prec, and returns the (possibly) rounded
// value of d according to the given rounding mode.
func (d *Decimal) ChangePrec(prec uint, rounding Rounding) *Decimal {
	return d
}

// Cmp compares d and x, and returns:
//
//     -1 if d < x
//     0 if d == x
//     +1 if d > x
func (d *Decimal) Cmp(x *Decimal) int {
	return 0
}

// Copy sets d to x, and returns d.
func (d *Decimal) Copy(x *Decimal) *Decimal {
	d.base = x.base
	d.excess = x.excess
	d.state = x.state
	return d
}

// Div sets d to the division of x by y, and returns d. It returns an error if y
// is zero, will use a max scale of 20, and if rounding is needed, will use
// ToNearestEven rounding.
func (d *Decimal) Div(x *Decimal, y *Decimal) (*Decimal, error) {
	return d, nil
}

// DivExplicit behaves the same as Div, except with a specific max scale and
// rounding mode.
func (d *Decimal) DivExplicit(x *Decimal, y *Decimal, scale uint, rounding Rounding) (*Decimal, error) {
	return d, nil
}

// Format implements the interface for fmt.Formatter.
func (d *Decimal) Format(s fmt.State, format rune) {
}

// GetRaw returns the components that represent the raw value of d.
func (d *Decimal) GetRaw() (base uint64, excess *big.Int, state uint64) {
	return d.base, &d.excess, d.state
}

// Int64 returns the integer resulting from truncating d towards zero. If the
// resulting value would be out of bounds for an int64, then ok will be false.
func (d *Decimal) Int64() (v int64, ok bool) {
	return 0, false
}

// IsInt returns whether the value is an integer.
func (d *Decimal) IsInt() bool {
	return false
}

// Mul sets d to the product of x and y, and returns d.
func (d *Decimal) Mul(x *Decimal, y *Decimal) *Decimal {
	return d
}

// Neg sets d to the negated value of x, and returns d.
func (d *Decimal) Neg(x *Decimal) *Decimal {
	return d
}

// Round rescales d to the given maximum scale, and returns the rounded value of
// d according to the given rounding mode.
func (d *Decimal) Round(scale uint, rounding Rounding) *Decimal {
	return d
}

// SetInt64 sets d's value to v, and returns d.
func (d *Decimal) SetInt64(v int64) *Decimal {
	return d
}

// SetRaw sets d to the value signified by the raw components.
func (d *Decimal) SetRaw(base uint64, excess big.Int, state uint64) {
	d.base = base
	d.excess = excess
	d.state = state
}

// SetString sets d to the parsed value to v in the given base, and returns d.
func (d *Decimal) SetString(v string, base uint) (*Decimal, error) {
	return d, nil
}

// SetUint64 sets d's value to v, and returns d.
func (d *Decimal) SetUint64(v uint64) *Decimal {
	return d
}

// Sign returns:
//
//     -1 if d < 0
//     0 if d == 0
//     +1 if d > 0
func (d *Decimal) Sign() int {
	return 0
}

// Sqrt sets d to the square root of x, and returns d. It returns an error if x
// is negative, will use a max scale of 20, and if rounding is needed, will use
// ToNearestEven rounding.
func (d *Decimal) Sqrt(x *Decimal) (*Decimal, error) {
	return d, nil
}

// SqrtExplicit behaves the same as Sqrt, except with a specific max scale and
// rounding mode.
func (d *Decimal) SqrtExplicit(x *Decimal, scale uint, rounding Rounding) (*Decimal, error) {
	return d, nil
}

// String returns a decimal representation of d. It will use a max scale of 20,
// and if rounding is needed, will use ToNearestEven rounding.
func (d *Decimal) String() string {
	return ""
}

// StringExplicit behaves the same as String, except with a specific max scale
// and rounding mode.
func (d *Decimal) StringExplicit(scale uint, rounding Rounding) string {
	return ""
}

// Sub sets d to the difference between x and y, and returns d.
func (d *Decimal) Sub(x *Decimal, y *Decimal) *Decimal {
	return d
}

// Uint64 returns the unsigned integer resulting from truncating d towards zero.
// If the resulting value would be out of bounds for a uint64, then ok will be
// false.
func (d *Decimal) Uint64() (v uint64, ok bool) {
	return 0, false
}

// Rounding specifies the rounding mode for certain operations.
type Rounding int

// FromInt64 returns a decimal with the value set to v.
func FromInt64(v int64) *Decimal {
	return New().SetInt64(v)
}

// FromRaw returns a decimal with the value signified by the raw components.
func FromRaw(base uint64, excess big.Int, state uint64) *Decimal {
	return &Decimal{
		base:   base,
		excess: excess,
		state:  state,
	}
}

// FromString returns a decimal with the parsed value of v in the given base.
func FromString(v string, base uint) (*Decimal, error) {
	return New().SetString(v, base)
}

// FromUint64 returns a decimal with the value set to v.
func FromUint64(v uint64) *Decimal {
	return New().SetUint64(v)
}

// MustDecimal returns a decimal value from the given decimal string
// representation. It panics if there was an error parsing the string.
func MustDecimal(v string) *Decimal {
	d, err := FromString(v, 10)
	if err != nil {
		panic(err)
	}
	return d
}

// New returns a zero valued decimal.
func New() *Decimal {
	return &Decimal{}
}
