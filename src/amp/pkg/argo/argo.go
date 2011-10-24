// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	"amp/big"
)

const (
	String = iota
	Int
	Int64
	True
	False
	StringSlice
	Dict
	Header
	ByteSlice
	Slice
	BigDecimal
	BigInt
)

const (
	magicNumber int64 = 8258175
)

var (
	bigintMagicNumber1, _ = big.NewIntString("8258175")
	bigintMagicNumber2, _ = big.NewIntString("8323072")
	bigint1, _            = big.NewIntString("1")
	bigint253, _          = big.NewIntString("253")
	bigint254, _          = big.NewIntString("254")
	bigint255, _          = big.NewIntString("255")
	zero                  = []byte{'\x01', '\x80', '\x01', '\x01'}
	zeroBase              = []byte{'\x80', '\x01', '\x01'}
)

type Error string

func (err Error) String() string {
	return string(err)
}

