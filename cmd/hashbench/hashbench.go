// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Command hashbench does a basic benchmark of various hash functions.
package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"math/rand"
	"time"

	"dappui.com/pkg/bytesize"
	"github.com/mimoo/GoKangarooTwelve/K12"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
	"lukechampine.com/blake3"
)

// XOF represents an eXtendable-Output Function.
type XOF interface {
	io.Reader
	io.Writer
}

func execHash(hash hash.Hash, data []byte) ([]byte, error) {
	n, err := hash.Write(data)
	if err != nil {
		return nil, err
	}
	if n != len(data) {
		return nil, fmt.Errorf("failed to hash all data, only %d bytes", n)
	}
	digest := hash.Sum(nil)
	return digest, nil
}

func execXOF(xof XOF, data []byte) ([]byte, error) {
	n, err := xof.Write(data)
	if err != nil {
		return nil, err
	}
	if n != len(data) {
		return nil, fmt.Errorf("failed to hash all data, only %d bytes", n)
	}
	digest := make([]byte, 64)
	n, err = xof.Read(digest)
	if err != nil {
		return nil, err
	}
	if n != 64 {
		return nil, fmt.Errorf("failed to read all data, only %d bytes", n)
	}
	return digest[32:], nil
}

func timeHash(name string, f func([]byte) ([]byte, error), data []byte) {
	var (
		digest []byte
		err    error
	)
	best := time.Duration(-1)
	for i := 0; i < 10; i++ {
		start := time.Now()
		digest, err = f(data)
		if err != nil {
			fmt.Printf("!! Failed to hash %s: %s\n", name, err)
			return
		}
		taken := time.Since(start)
		if best == -1 || taken < best {
			best = taken
		}
	}
	fmt.Printf(">> Time taken for %-15s %15s (%x)\n", name+":", best, digest)
}

func runBlake2b(data []byte) ([]byte, error) {
	hash, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}
	return execHash(hash, data)
}

func runBlake2s(data []byte) ([]byte, error) {
	hash, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}
	return execHash(hash, data)
}

func runBlake3(data []byte) ([]byte, error) {
	return execHash(blake3.New(32, nil), data)
}

func runK12(data []byte) ([]byte, error) {
	hash := K12.NewK12(nil)
	n, err := hash.Write(data)
	if err != nil {
		return nil, err
	}
	if n != len(data) {
		return nil, fmt.Errorf("failed to hash all data, only %d bytes", n)
	}
	digest := make([]byte, 32)
	n, err = hash.Read(digest)
	if err != nil {
		return nil, err
	}
	if n != 32 {
		return nil, fmt.Errorf("failed to read all data, only %d bytes", n)
	}
	return digest, nil
}

func runMD5(data []byte) ([]byte, error) {
	return execHash(md5.New(), data)
}

func runSHA1(data []byte) ([]byte, error) {
	return execHash(sha1.New(), data)
}

func runSHA2_256(data []byte) ([]byte, error) {
	return execHash(sha256.New(), data)
}

func runSHA2_512_256(data []byte) ([]byte, error) {
	return execHash(sha512.New512_256(), data)
}

func runSHA3_256(data []byte) ([]byte, error) {
	return execHash(sha3.New256(), data)
}

func runSHAKE256(data []byte) ([]byte, error) {
	return execXOF(sha3.NewShake256(), data)
}

func main() {
	for _, size := range []bytesize.Value{
		100,
		bytesize.KB,
		10 * bytesize.MB,
		100 * bytesize.MB,
	} {
		data := make([]byte, size.MustInt())
		rand.Seed(123456789)
		rand.Read(data)
		fmt.Printf("## Running benchmark for data of size %s\n\n", size)
		timeHash("blake2b", runBlake2b, data)
		timeHash("blake2s", runBlake2s, data)
		timeHash("blake3", runBlake3, data)
		timeHash("k12", runK12, data)
		timeHash("md5", runMD5, data)
		timeHash("sha1", runSHA1, data)
		timeHash("sha2-256", runSHA2_256, data)
		timeHash("sha2-512/256", runSHA2_512_256, data)
		timeHash("sha3-256", runSHA3_256, data)
		timeHash("shake256", runSHAKE256, data)
		fmt.Println("")
	}
}
