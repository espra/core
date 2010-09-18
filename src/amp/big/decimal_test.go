// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package big

import (
	"testing"
	"fmt"
)

func isValidParse(value, expected string) bool {
	dec, ok := NewDecimal(value)
	if !ok {
		return false
	}
	if dec.String() != expected {
		return false
	}
	return true
}

func TestDecimal(t *testing.T) {

	tests := map[string]string{
		"0.1":       "0.1",
		"0.":        "0",
		"-0.1":      "-0.1",
		"+0.1":      "0.1",
		"-0.0000":   "0",
		"0.0123107": "0.0123107",
		".0123107":  "0.0123107",
	}

	for value, expected := range tests {
		if !isValidParse(value, expected) {
			dec, ok := NewDecimal(value)
			t.Errorf("Decimal value did not match expected for %q\n", value)
			if ok {
				t.Errorf("Got: \"%v\"\nExpected: \"%s\"", dec, expected)
			} else {
				t.Error("Got not ok from `NewDecimal`\n")
			}
			return
		}
	}

}

func TestDecimalAdd(t *testing.T) {

	x, _ := NewDecimal(".0123107")
	y, _ := NewDecimal(".0123107")
	z, _ := NewDecimal("0")
	z = z.Add(x, y)
	fmt.Printf("z: %v\n", z)

	x, _ = NewDecimal(".0123107")
	y, _ = NewDecimal(".0123107")
	z, _ = NewDecimal("0")
	z = z.Sub(x, y)
	fmt.Printf("z: %v\n", z)

	x, _ = NewDecimal(".0123107")
	y, _ = NewDecimal(".0123107")
	z, _ = NewDecimal("0")
	z = z.Sub(z.Add(x, y), x)
	fmt.Printf("z: %v\n", z)

}

func BenchmarkDecimal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewDecimal("2.0123107")
	}
}

func BenchmarkRat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		new(Rat).SetString("2.0123107")
	}
}
