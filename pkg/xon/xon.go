// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Package xon implements the eXtensible Object Notation (XON) format.
package xon

import (
	"fmt"
	"math/big"
	"strconv"
)

var floatZero = big.NewFloat(0)

// IsZeroer defines the interface that struct values should implement so that
// the marshaler can determine whether to omit the value.
type IsZeroer interface {
	IsZero() bool
}

// Marshaler is the interface implemented by types that can marshal themselves
// into valid XON.
type Marshaler interface {
	MarshalXON() ([]byte, error)
}

// Number represents a arbitrary-precision number.
type Number struct {
	raw string
}

// BigFloat tries to parse a big.Float value from the number. The returned value
// is set to the equivalent precision of a binary128 value, i.e. 113 bits, and
// has the rounding mode set to nearest even.
func (n Number) BigFloat() (*big.Float, error) {
	if n.raw == "" {
		return &big.Float{}, nil
	}
	v, _, err := big.ParseFloat(n.raw, 10, 113, big.ToNearestEven)
	return v, err
}

// BigInt tries to parse a big.Int value from the number.
func (n Number) BigInt() (*big.Int, error) {
	if n.raw == "" {
		return &big.Int{}, nil
	}
	v, ok := new(big.Int).SetString(n.raw, 10)
	if !ok {
		return nil, fmt.Errorf("xon: unable to parse %q into a big.Int", n.raw)
	}
	return v, nil
}

// Float64 tries to parse a float64 value from the number.
func (n Number) Float64() (float64, error) {
	if n.raw == "" {
		return 0, nil
	}
	return strconv.ParseFloat(n.raw, 64)
}

// Int64 tries to parse an int64 value from the number.
func (n Number) Int64() (int64, error) {
	if n.raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(n.raw, 10, 64)
}

// IsZero implements the IsZeroer interface.
func (n Number) IsZero() bool {
	if n.raw == "" {
		return true
	}
	f, err := n.BigFloat()
	if err != nil {
		panic(fmt.Errorf("xon: failed to convert %q into a big.Float: %s", n.raw, err))
	}
	return f.Cmp(floatZero) == 0
}

// MarshalXON implements the Marshaler interface.
func (n Number) MarshalXON() ([]byte, error) {
	return []byte(n.raw), nil
}

func (n Number) String() string {
	if n.raw == "" {
		return "0"
	}
	return n.raw
}

// Uint64 tries to parse a uint64 value from the number.
func (n Number) Uint64() (uint64, error) {
	if n.raw == "" {
		return 0, nil
	}
	return strconv.ParseUint(n.raw, 10, 64)
}

// UnmarshalXON implements the Marshaler interface.
func (n *Number) UnmarshalXON(data []byte) error {
	if len(data) == 0 {
		n.raw = ""
		return nil
	}
	// TODO(tav): Perhaps validate the number.
	n.raw = string(data)
	return nil
}

// UnitValue represents a numeric component followed by a unit.
type UnitValue struct {
	Unit  string
	Value Number
}

// IsZero implements the IsZeroer interface.
func (u UnitValue) IsZero() bool {
	return u.Unit == "" && u.Value.IsZero()
}

func (u UnitValue) String() string {
	return u.Value.String() + u.Unit
}

// Unmarshaler is the interface implemented by types that can unmarshal a XON
// description of themselves. The input can be assumed to be a valid encoding of
// a XON value. UnmarshalXON must copy the XON data if it wishes to retain the
// data after returning.
type Unmarshaler interface {
	UnmarshalXON([]byte) error
}

// Variant represents the name and fields of an enum variant.
type Variant struct {
	Fields map[string]interface{}
	Name   string
}

// IsZero implements the IsZeroer interface.
func (v Variant) IsZero() bool {
	return len(v.Fields) == 0 && v.Name == ""
}

// Format will reformat the given src in the canonical XON style.
func Format(src []byte) ([]byte, error) {
	return nil, nil
}

// Marshal returns the XON encoding of v.
func Marshal(v interface{}) ([]byte, error) {
	return nil, nil
}

// RegisterVariant allows for the registration of an enum variant. The given typ
// must be a pointer to an interface, and the variant must be a pointer to a
// struct implementing that interface. The name of the struct type will be used
// when the variant is converted to/from XON.
func RegisterVariant(typ interface{}, variant interface{}) error {
	return nil
}

// Unmarshal parses the XON-encoded data and stores the result in the value
// pointed to by v.
//
// When unmarshalling into an empty interface value, the following types are
// used:
//
//     []byte, for XON binary blobs
//     bool, for XON booleans
//     bytesize.Value, for XON bytesize values
//     Duration, for XON duration values
//     Number, for XON numbers
//     string, for XON strings
//     time.Time, for XON dates and timestamps
//     UnitValue, for XON unit values
//     []interface{}, for XON lists
//     map[string]interface{}, for XON structs
//     Variant, for XON enum variants
//
func Unmarshal(data []byte, v interface{}) error {
	return nil
}

// UnmarshalStrict behaves the same as Unmarshal, except that it will error when
// decoding a field into a struct that doesn't have a field with the same name.
func UnmarshalStrict(data []byte, v interface{}) error {
	return nil
}
