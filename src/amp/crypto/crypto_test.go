// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package crypto

import (
	"bytes"
	"crypto/hmac"
	"encoding/hex"
	"strings"
	"testing"
)

type pbkdf2 struct {
	password   string
	salt       string
	iterations int
	keysize    int
	expected   string
}

func hex2bytes(value string) []byte {
	split := strings.Split(value, " ")
	output := make([]byte, len(split))
	for i, octet := range split {
		decoded, _ := hex.DecodeString(octet)
		output[i] = decoded[0]
	}
	return output
}

// We test the PBKDF2 implementation against the SHA-1 test vectors defined in
// `RFC 6070 <http://www.ietf.org/rfc/rfc6070.txt>`_.
func TestPBKDF2_SHA1(t *testing.T) {

	tests := []pbkdf2{
		{"password", "salt", 1, 20, "0c 60 c8 0f 96 1f 0e 71 f3 a9 b5 24 af 60 12 06 2f e0 37 a6"},
		{"password", "salt", 2, 20, "ea 6c 01 4d c7 2d 6f 8c cd 1e d9 2a ce 1d 41 f0 d8 de 89 57"},
		{"password", "salt", 4096, 20, "4b 00 79 01 b7 65 48 9a be ad 49 d9 26 f7 21 d0 65 a4 29 c1"},
		{"passwordPASSWORDpassword", "saltSALTsaltSALTsaltSALTsaltSALTsalt", 4096, 25,
			"3d 2e ec 4f e4 1c 84 9b 80 c8 d8 36 62 c0 e4 4a 8b 29 1a 96 4c f2 f0 70 38"},
		{"pass\x00word", "sa\x00lt", 4096, 16, "56 fa 6a a7 55 48 09 9d cc 37 d7 f0 34 25 e0 c3"},
	}

	// Left out this test vector as it takes way too long...
	//
	//	{"password", "salt", 16777216, 20,
	//	    "ee fe 3d 61 cd 4d a4 e4 e9 94 5b 3d 6b a2 15 8c 26 34 e9 84"},

	for idx, test := range tests {
		output := PBKDF2(hmac.NewSHA1, []byte(test.password), []byte(test.salt), test.iterations, test.keysize)
		expected := hex2bytes(test.expected)
		if !bytes.Equal(output, expected) {
			t.Errorf("PBKDF2-SHA1 mismatch for test #%d: %#v\n", idx+1, test)
		}
	}

}

// We test the PBKDF2 implementation against the SHA-256 test vectors posted at
// <http://stackoverflow.com/questions/5130513/pbkdf2-hmac-sha2-test-vectors>.
func TestPBKDF2_SHA256(t *testing.T) {

	tests := []pbkdf2{
		{"password", "salt", 1, 20, "12 0f b6 cf fc f8 b3 2c 43 e7 22 52 56 c4 f8 37 a8 65 48 c9"},
		{"password", "salt", 2, 20, "ae 4d 0c 95 af 6b 46 d3 2d 0a df f9 28 f0 6d d0 2a 30 3f 8e"},
		{"password", "salt", 4096, 20, "c5 e4 78 d5 92 88 c8 41 aa 53 0d b6 84 5c 4c 8d 96 28 93 a0"},
		{"passwordPASSWORDpassword", "saltSALTsaltSALTsaltSALTsaltSALTsalt", 4096, 25,
			"34 8c 89 db cb d3 2b 2f 32 d8 14 b8 11 6e 84 cf 2b 17 34 7e bc 18 00 18 1c"},
		{"pass\x00word", "sa\x00lt", 4096, 16, "89 b6 9d 05 16 f8 29 89 3c 69 62 26 65 0a 86 87"},
	}

	// Left out this test vector as it takes way too long...
	//
	//	{"password", "salt", 16777216, 20,
	//	    "cf 81 c6 6f e8 cf c0 4d 1f 31 ec b6 5d ab 40 89 f7 f1 79 e8"},

	for idx, test := range tests {
		output := PBKDF2(hmac.NewSHA256, []byte(test.password), []byte(test.salt), test.iterations, test.keysize)
		expected := hex2bytes(test.expected)
		if !bytes.Equal(output, expected) {
			t.Errorf("PBKDF2-SHA256 mismatch for test #%d: %#v\n", idx+1, test)
		}
	}

}

func TestPassword(t *testing.T) {
	password, err := NewPassword("letmein")
	if err != nil {
		t.Errorf("Got an error reading random source: %s", err)
		return
	}
	invalid := "not the right password"
	if password.Validate(invalid) {
		t.Errorf("Got a validation for invalid password %q: %#v", invalid, password)
		return
	}
	if !password.Validate("letmein") {
		t.Errorf("Didn't validate the right password in %#v", password)
		return
	}
}

func TestIronStrings(t *testing.T) {

	key := []byte("sekret key")
	iron := IronString("user", "tav", key, 60)
	value, ok := GetIronValue("user", iron, key, true)

	if !ok {
		t.Errorf("Got an error getting the value for the IronString.")
	}

	if value != "tav" {
		t.Errorf("Got an invalid value for the IronString: %q", value)
	}

	value, ok = GetIronValue("user", iron, []byte("wrong key"), true)
	if ok {
		t.Errorf("Got an ok with the wrong key for the IronString.")
	}

	tampered := "X" + iron[1:]

	value, ok = GetIronValue("user", tampered, key, true)
	if ok {
		t.Errorf("Got an ok with a tampered IronString.")
	}

	iron = IronString("user", "tav", key, -60)

	value, ok = GetIronValue("user", iron, key, true)
	if ok {
		t.Errorf("Got an ok for an outdated IronString.")
	}

	iron = IronString("age", "29", key, 0)

	value, ok = GetIronValue("age", iron, key, true)
	if ok {
		t.Errorf("Got an ok for a non-timestamped IronString claiming to be timestamped.")
	}

	value, ok = GetIronValue("age", iron, key, false)
	if !ok {
		t.Errorf("Got an error validating a non-timestamped IronString.")
	}

	if value != "29" {
		t.Errorf("Got an invalid value for the IronString: %q", value)
	}

	iron = IronString("user|location", "london", key, 0)

	value, ok = GetIronValue("user|location", iron, key, false)
	if !ok || value != "london" {
		t.Errorf("Got an error validating an IronString with a | in the property name.")
	}

	iron = IronString("user|location", "london", key, 60)

	value, ok = GetIronValue("user|location", iron, key, true)
	if !ok || value != "london" {
		t.Errorf("Got an error validating a timestamped IronString with a | in the property name.")
	}

}
