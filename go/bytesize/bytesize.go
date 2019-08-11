// Public Domain (-) 2018-present, The Amp Authors.
// See the Amp UNLICENSE file for details.

// Package bytesize provides support for dealing with byte size values.
package bytesize

import (
	"fmt"
	"strconv"
	"strings"

	"ampify.dev/go/overflow"
)

// Constants representing common byte sizes.
const (
	B    Value = 1
	Byte       = 1
	KB         = 1024
	MB         = 1024 * KB
	GB         = 1024 * MB
	TB         = 1024 * GB
	PB         = 1024 * TB
)

const maxInt = Value(^uint(0) >> 1)

// Value represents a byte size value.
type Value uint64

// Int checks if the value would overflow the platform int and, if not, returns
// the value as an int.
func (v Value) Int() (int, error) {
	if v > maxInt {
		return 0, fmt.Errorf("bytesize: value %d (%s) overflows platform int", v, v.String())
	}
	return int(v), nil
}

// String produces a human-readable representation of the byte size value as a
// sequence of decimal numbers followed by a unit suffix. Where possible, the
// unit yielding the smallest possible string representation will be chosen.
//
// Valid units are "B", "KB", "MB", "GB", "TB", and "PB", with "B" being the
// default where the byte size value is not an exact multiple of any of the
// larger units.
func (v Value) String() string {
	switch {
	case v%PB == 0:
		return strconv.FormatUint(uint64(v/PB), 10) + "PB"
	case v%TB == 0:
		return strconv.FormatUint(uint64(v/TB), 10) + "TB"
	case v%GB == 0:
		return strconv.FormatUint(uint64(v/GB), 10) + "GB"
	case v%MB == 0:
		return strconv.FormatUint(uint64(v/MB), 10) + "MB"
	case v%KB == 0:
		return strconv.FormatUint(uint64(v/KB), 10) + "KB"
	default:
		return strconv.FormatUint(uint64(v), 10) + "B"
	}
}

// Parse tries to parse a byte size value from the given string. A byte size
// string is a sequence of decimal numbers and a unit suffix, e.g. "20GB",
// "1024KB", "100MB", etc. Valid units are "B", "KB", "MB", "GB", "TB", and
// "PB". If no unit is specified, then the value is assumed to be in bytes.
func Parse(s string) (Value, error) {
	var (
		err  error
		ok   bool
		unit string
		v    uint64
	)
	for i := len(s) - 1; i >= 0; i-- {
		char := s[i]
		if char >= '0' && char <= '9' {
			v, err = strconv.ParseUint(s[:i+1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("bytesize: unable to parse the decimal part of %q: %s", s, err)
			}
			unit = s[i+1:]
			break
		}
	}
	switch strings.ToLower(unit) {
	case "", "b":
		ok = true
	case "kb":
		v, ok = overflow.MulU64(v, uint64(KB))
	case "mb":
		v, ok = overflow.MulU64(v, uint64(MB))
	case "gb":
		v, ok = overflow.MulU64(v, uint64(GB))
	case "tb":
		v, ok = overflow.MulU64(v, uint64(TB))
	case "pb":
		v, ok = overflow.MulU64(v, uint64(PB))
	default:
		return 0, fmt.Errorf("bytesize: unsupported unit %q specified in %q", unit, s)
	}
	if !ok {
		return 0, fmt.Errorf("bytesize: string value %q overflows uint64 when parsed", s)
	}
	return Value(v), nil
}
