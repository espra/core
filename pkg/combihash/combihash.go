// Public Domain (-) 2018-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Package combihash implements the combined hash format.
//
// Combihash digests are good as identifiers in content addressable systems. The
// digests are built from two different hash functions, so if any one of them
// gets broken, you are still protected by the security properties of the other.
//
// The digest format is composed of:
//
//     <hash1-digest> <hash2-digest> <data-length> <version>
//     <X bytes>      <Y bytes>      <7 bytes>     <1 byte>
//
// Where X and Y represents the respective lengths of the digests from hash1 and
// hash2 for a particular version of Combihash.
//
// Combihash can produce digests for data of lengths up to 2^56 - 1 bytes of
// data. The digests include a version byte, so applications can concurrently
// support newer versions while older versions are phased out.
//
// The digest format includes the most distinctive bytes at the start so as to
// support the use of short unique prefixes in applications, e.g. like in Git.
//
// Version 1 of Combihash is built around BLAKE3 and Kangaroo12, and the
// generated digest is 72 bytes long:
//
//     <blake3-digest> <kangaroo12-digest> <data-length> <version>
//     <32 bytes>      <32 bytes>          <7 bytes>     <1 byte>
//
package combihash

import (
	"errors"

	"dappui.com/pkg/blake3"
	"dappui.com/pkg/crypto"
	"dappui.com/pkg/kangaroo12"
)

// MaxLength specifies the maximum length of data that can be hashed using
// Combihash.
const MaxLength = (1 << 56) - 1

var (
	errExceedsMaxLength = errors.New("combihash: data exceeds max length of 2^56 - 1 bytes")
)

// Digest enables inspecting the components of a Combihash digest. The digest
// should be checked with a call to IsValid first, otherwise the other methods
// may panic.
type Digest []byte

// IsValid returns whether it's a valid digest, i.e. of the correct length and
// with a valid version identifier.
func (d Digest) IsValid() bool {
	return len(d) == 72 && d[71] == 1
}

// Length returns the length of the hashed data.
func (d Digest) Length() uint64 {
	_ = d[70] // bounds check hint to compiler
	return uint64(d[64]) | uint64(d[65])<<8 | uint64(d[66])<<16 | uint64(d[67])<<24 |
		uint64(d[68])<<32 | uint64(d[69])<<40 | uint64(d[70])<<48
}

// Primary returns the digest component from the primary hash.
func (d Digest) Primary() []byte {
	return d[:32]
}

// Secondary returns the digest component from the secondary hash.
func (d Digest) Secondary() []byte {
	return d[32:64]
}

// Version returns the Combihash version of the digest.
func (d Digest) Version() int {
	return int(d[71])
}

type digester struct {
	blake3     crypto.Digester
	kangaroo12 crypto.Digester
}

func (d *digester) Digest(data []byte, b []byte) []byte {
	if len(data) > MaxLength {
		panic(errExceedsMaxLength)
	}
	var digest []byte
	blen := len(b)
	if n := blen + 72; cap(b) >= n {
		digest = b[:n]
	} else {
		digest = make([]byte, n)
		copy(digest, b)
	}
	d.blake3.Digest(data, digest[blen:blen])
	d.kangaroo12.Digest(data, digest[blen+32:blen+32])
	setLength(digest[blen+64:], uint64(len(data)))
	digest[71] = 1
	return digest
}

// New instantiates a crypto.Digester for version 1 of Combihash.
func New() crypto.Digester {
	return &digester{
		blake3:     blake3.NewDigester(),
		kangaroo12: kangaroo12.NewDigester(),
	}
}

func setLength(b []byte, v uint64) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
}
