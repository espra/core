// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package big

import (
	"testing"
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
		"5":         "5",
		"1":         "1",
		"-5":        "-5",
		"10":        "10",
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

func TestDecimalCopy(t *testing.T) {

	dec, _ := NewDecimal("0.01")
	if dec.Copy().String() != "0.01" {
		t.Error("Copying `NewDecimal(\"0.01\")` failed.")
		t.Errorf("Got \"%s\".\n", dec.Copy().String())
	}

}

func TestDecimalInvalids(t *testing.T) {

	tests := []string{
		"0.012.3107",
		"invalid",
		"decafbad",
	}

	for _, value := range tests {
		if _, ok := NewDecimal(value); ok {
			t.Errorf("Got a valid parse for %q\n", value)
			return
		}
	}

}

func decimal(value string) *Decimal {
	dec, _ := NewDecimal(value)
	return dec
}

func TestDecimalAdd(t *testing.T) {

	tests := [][]string{
		[]string{"0.1", "0.2", "0.3"},
		[]string{"-0.1", "0.2", "0.1"},
		[]string{"-0.1", "-0.2", "-0.3"},
	}

	for _, test := range tests {
		x := test[0]
		y := test[1]
		expected := test[2]
		result := decimal(x).Add(decimal(y))
		if result.String() != expected {
			t.Errorf("Expected: %s + %s = %s\n", x, y, expected)
			t.Errorf("Got %s\n", result.String())
			return
		}
	}

}

func TestDecimalSub(t *testing.T) {

	tests := [][]string{
		[]string{"0.1", "0.2", "-0.1"},
		[]string{"-0.1", "0.2", "-0.3"},
		[]string{"-0.1", "-0.2", "0.1"},
		[]string{"0.1", "0.1", "0"},
	}

	for _, test := range tests {
		x := test[0]
		y := test[1]
		expected := test[2]
		result := decimal(x).Sub(decimal(y))
		if result.String() != expected {
			t.Errorf("Expected: %s - %s = %s\n", x, y, expected)
			t.Errorf("Got %s\n", result.String())
			return
		}
	}

}

func TestDecimalMul(t *testing.T) {

	tests := [][]string{
		[]string{"0.1", "0.2", "0.02"},
		[]string{"-0.1", "0.2", "-0.02"},
		[]string{"-0.1", "-0.2", "0.02"},
	}

	for _, test := range tests {
		x := test[0]
		y := test[1]
		expected := test[2]
		result := decimal(x).Mul(decimal(y))
		if result.String() != expected {
			t.Errorf("Expected: %s * %s = %s\n", x, y, expected)
			t.Errorf("Got %s\n", result.String())
			return
		}
	}

}

func TestDecimalDiv(t *testing.T) {

	tests := [][]string{
		[]string{"18", "18", "1"},
		[]string{"18", "-18", "-1"},
		[]string{"0.5", "0.2", "2.5"},
		[]string{"0.5", "-0.5", "-1"},
		[]string{"4", "-3", "-1.3333333333333333333333333333333333333333"},
		[]string{"0.000000000000000000000000000000000000000000001", "20000000000",
			"0"},
		[]string{"0.0000000000000000000000000000000000000001", "1",
			"0.0000000000000000000000000000000000000001"},
		[]string{"0.000000000000000000000000000000000000000000001", "1",
			"0"},
		[]string{"4", "-0.000000000000000000000000000000000000000000001",
			"-4000000000000000000000000000000000000000000000"},
		[]string{"4", "-0.00000000000000000000000000000000000000001",
			"-400000000000000000000000000000000000000000"},
		[]string{"4.51", "-3.21", "-1.4049844236760124610591900311526479750778"},
		[]string{"-3", "-4", "0.75"},
		[]string{"-0.1", "0.2", "-0.5"},
		[]string{"-0.1", "-0.2", "0.5"},
	}

	for _, test := range tests {
		x := test[0]
		y := test[1]
		expected := test[2]
		result := decimal(x).Div(decimal(y))
		if result.String() != expected {
			t.Errorf("Expected: %s / %s = %s\n", x, y, expected)
			t.Errorf("Got       %s\n", result.String())
			return
		}
	}

}

func compare(t *testing.T, tests [][]string, expected int) {

	for _, test := range tests {
		x := test[0]
		y := test[1]
		result := decimal(x).Cmp(decimal(y))
		if result != expected {
			t.Errorf("Expected: %q.Cmp(%q) = %d\n", x, y, expected)
			t.Errorf("Got %d\n", result)
			return
		}
	}

}

func TestDecimalCmp(t *testing.T) {

	testGreater := [][]string{
		[]string{"2", "1"},
		[]string{"0.500001", "0.5"},
		[]string{"-0.1", "-0.2"},
		[]string{"0.1", "-0.1"},
	}

	testLesser := [][]string{
		[]string{"2", "3"},
		[]string{"0.1", "0.1001"},
		[]string{"-0.1", "0.1"},
		[]string{"-0.2", "-0.1"},
	}

	testEqual := [][]string{
		[]string{"1", "1"},
		[]string{"0.1", "0.1"},
		[]string{"-0.1", "-0.1"},
		[]string{"-0.100", "-000000.100000000000000"},
	}

	compare(t, testGreater, 1)
	compare(t, testLesser, -1)
	compare(t, testEqual, 0)

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
