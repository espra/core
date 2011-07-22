// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"hash"
	"os"
	"strconv"
	"strings"
	"time"
)

// -----------------------------------------------------------------------------
// PBKDF2
// -----------------------------------------------------------------------------

// An implementation of PBKDF2 (Password-Based Key Derivation Function 2) as
// specified in PKCS #5 v2.0 from RSA Laboratorie and in `RFC 2898
// <http://www.ietf.org/rfc/rfc2898.txt>`.
func PBKDF2(hashfunc func([]byte) hash.Hash, password, salt []byte, iterations, keylen int) (key []byte) {

	var (
		digest          []byte
		i, j, k, length int
	)

	key = make([]byte, keylen)
	slice := key

	hash := hashfunc(password)
	hashlen := hash.Size()
	scratch := make([]byte, 4)

	for keylen > 0 {

		if hashlen > keylen {
			length = keylen
		} else {
			length = hashlen
		}

		i += 1

		scratch[0] = byte(i >> 24)
		scratch[1] = byte(i >> 16)
		scratch[2] = byte(i >> 8)
		scratch[3] = byte(i)

		hash.Write(salt)
		hash.Write(scratch)

		digest = hash.Sum()
		hash.Reset()

		for j = 0; j < length; j++ {
			slice[j] = digest[j]
		}

		for k = 1; k < iterations; k++ {
			hash.Write(digest)
			digest = hash.Sum()
			for j = 0; j < length; j++ {
				slice[j] ^= digest[j]
			}
			hash.Reset()
		}

		keylen -= length
		slice = slice[length:]

	}

	return

}

// -----------------------------------------------------------------------------
// Password Generator Controls
// -----------------------------------------------------------------------------

var (
	passwordHashFunc   = hmac.NewSHA256
	passwordIterations = 20000
	passwordKeyLength  = 80
	passwordSaltLength = 10
)

func SetPasswordHashFunc(hashfunc func([]byte) hash.Hash) {
	passwordHashFunc = hashfunc
}

func SetPasswordIterations(n int) {
	passwordIterations = n
}

func SetPasswordKeyLength(n int) {
	passwordKeyLength = n
}

func SetPasswordSaltLength(n int) {
	passwordSaltLength = n
}

// -----------------------------------------------------------------------------
// PBKDF2-Based Password Support
// -----------------------------------------------------------------------------

type Password struct {
	Iterations int
	Key        []byte
	KeyLength  int
	Salt       []byte
}

func NewPassword(secret string) (password *Password, err os.Error) {
	salt := make([]byte, passwordSaltLength)
	_, err = rand.Read(salt)
	if err != nil {
		return
	}
	return &Password{
		Iterations: passwordIterations,
		Key: PBKDF2(
			passwordHashFunc, []byte(secret), salt, passwordIterations, passwordKeyLength),
		KeyLength: passwordKeyLength,
		Salt:      salt,
	}, nil
}

func (password *Password) Validate(secret string) (valid bool) {
	return password.ValidateWithHashFunc(secret, passwordHashFunc)
}

func (password *Password) ValidateWithHashFunc(secret string, hashfunc func([]byte) hash.Hash) (valid bool) {
	expected := PBKDF2(hashfunc, []byte(secret), password.Salt, password.Iterations, password.KeyLength)
	if len(expected) != len(password.Key) {
		return
	}
	return subtle.ConstantTimeCompare(password.Key, expected) == 1
}

// -----------------------------------------------------------------------------
// Tamper-Resistant IronStrings
// -----------------------------------------------------------------------------

var ironHMAC = hmac.NewSHA256

func SetIronHMAC(hmac func([]byte) hash.Hash) {
	ironHMAC = hmac
}

func IronString(name, value string, key []byte, duration int64) string {
	if duration > 0 {
		value = fmt.Sprintf("%d:%s", time.Seconds()+duration, value)
	}
	message := fmt.Sprintf("%s|%s", strings.Replace(name, "|", `\|`, -1), value)
	hmac := ironHMAC(key)
	hmac.Write([]byte(message))
	mac := base64.URLEncoding.EncodeToString(hmac.Sum())
	return fmt.Sprintf("%s:%s", mac, value)
}

func GetIronValue(name, value string, key []byte, timestamped bool) (val string, ok bool) {
	split := strings.SplitN(value, ":", 2)
	if len(split) != 2 {
		return
	}
	expected, value := []byte(split[0]), split[1]
	message := fmt.Sprintf("%s|%s", strings.Replace(name, "|", `\|`, -1), value)
	hmac := ironHMAC(key)
	hmac.Write([]byte(message))
	digest := hmac.Sum()
	mac := make([]byte, base64.URLEncoding.EncodedLen(len(digest)))
	base64.URLEncoding.Encode(mac, digest)
	if subtle.ConstantTimeCompare(mac, expected) != 1 {
		return
	}
	if timestamped {
		split = strings.SplitN(value, ":", 2)
		if len(split) != 2 {
			return
		}
		timestring, val := split[0], split[1]
		timestamp, err := strconv.Atoi64(timestring)
		if err != nil {
			return
		}
		if time.Seconds() > timestamp {
			return
		}
		return val, true
	}
	return value, true
}
