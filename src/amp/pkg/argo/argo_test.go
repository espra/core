// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	// "amp/big"
	"bytes"
	"gob"
	"json"
	// "fmt"
	// "reflect"
	"testing"
	"time"
)

// func TestWriteSize(t *testing.T) {

// 	tests := map[int]string{
// 		0:         "\x00",
// 		123456789: "\x95\x9a\xef:",
// 	}

// 	for value, expected := range tests {
// 		buf := &bytes.Buffer{}
// 		enc := NewEncoder(buf)
// 		enc.WriteSize(value)
// 		if string(buf.Bytes()) != expected {
// 			t.Errorf("Got unexpected encoding for %d: %q", value, buf.Bytes())
// 		}
// 	}

// }

// func TestFloat32(t *testing.T) {

// 	tests := []float32{17.0, 1792.13, 21, 0.1233, 0, 1.1, 2.1, 3.1, 7.1, 8.1, 9.1, 10.1, 14.1, 15.1, 16.1, 20.1}

// 	for _, value := range tests {
// 		buf := &bytes.Buffer{}
// 		enc := NewEncoder(buf)
// 		enc.WriteFloat32(value)
// 	}

// }

func TestFloat64(t *testing.T) {

	// tests := []float64{17.0, 1792.13, 21, 0.1233, 0, 1.1, 2.1, 3.1, 7.1, 8.1, 9.1, 10.1, 14.1, 15.1, 16.1, 20.1}

	var output float64

	tests := []float64{17.0, 1792.13, 21, 0.1233, 0, 1.1, 2.1, 3.1, 7.1, 8.1, 9.1, 10.1, 14.1, 15.1, 16.1, 20.1}

	// output := make([]float64, 0)

	for _, value := range tests {
		buf := &bytes.Buffer{}
		enc := NewEncoder(buf)
		enc.Encode(value)
		dec := NewDecoder(buf)
		err := dec.Decode(&output)
		if err != nil {
			t.Logf("error: %v", err)
		}
		if output != value {
			t.Logf("got: %v, expected: %v", output, value)
		}
	}

}

type Crazy struct {
	name   string
	Age    int `argo:"CrazyAge";json:"CrazyAge";gob:"CrazyAge"`
	Likes  []string
	Height int
}

func TestCrazy(t *testing.T) {

	// var n uint64 = 18446744073709551615
	// var n uint64 = 2147483647
	// var n uint64 = 20
	// b := &bytes.Buffer{}
	// e := NewEncoder(b)
	// e.Encode(2)
	// t.Logf("%q", b.String())

	x := &Crazy{"tav", 29, []string{"nutella", "Gauloises"}, 180}

	value := []*Crazy{x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x}

	// value := []*Crazy{x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x}


	// x := map[string]interface{}{
	// 	"name":       "tav",
	// 	"age":        float64(29),
	// 	"housemates": []string{"Dave", "Mamading", "Sofia", "Sylvana"},
	// 	"meta": map[string]interface{}{
	// 		"City":     "London",
	// 		"Duration": 6,
	// 		"Month":    "March",
	// 		"Address": map[string]interface{}{
	// 			"Town":    "Brixton",
	// 			"Country": "U.K.",
	// 		},
	// 		"Sexy":     "jeffarch",
	// 		"Stylists": []string{"Phillipe", "Diane"},
	// 		"Birth":    float64(1982),
	// 	},
	// }
	// value := []map[string]interface{}{x, x, x, x, x}

	// value := []map[string]interface{}{x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x}

	gob.Register(x)
	gob.Register(value)
	// gob.Register(x)

	buf := &bytes.Buffer{}
	enc := NewEncoder(buf)
	start := time.Nanoseconds()
	enc.Encode(value)
	stop := time.Nanoseconds()
	output := buf.String()
	t.Logf("Time: %d", stop-start)
	t.Logf("%d", len(output))
	// t.Logf("%q", output)
	// t.Logf("%v", []byte(output))

	buf = &bytes.Buffer{}
	jenc := json.NewEncoder(buf)
	start = time.Nanoseconds()
	jenc.Encode(value)
	stop = time.Nanoseconds()
	output = buf.String()
	t.Logf("Time: %d", stop-start)
	t.Logf("%d", len(output))
	// t.Logf("%q", output)
	// t.Logf("%v", []byte(output))

	buf = &bytes.Buffer{}
	genc := gob.NewEncoder(buf)
	start = time.Nanoseconds()
	genc.Encode(value)
	stop = time.Nanoseconds()
	output = buf.String()
	// t.Logf("Err: %s", err)
	t.Logf("Time: %d", stop-start)
	t.Logf("%d", len(output))
	// t.Logf("%q", output)
	// t.Logf("%v", []byte(output))

	buf = &bytes.Buffer{}
	enc = NewEncoder(buf)
	start = time.Nanoseconds()
	enc.Encode(value)
	stop = time.Nanoseconds()
	output = buf.String()
	t.Logf("Time: %d", stop-start)
	t.Logf("%d", len(output))

}

// func testInt64(t *testing.T) {

// 	tests := []int64{170, 179213, 21, 01233, 0, 11, 21, 31, 71, 81, 91, 101, 141, 151, 161, 201, -1, 1, 2, -2}

// 	total := 0

// 	for _, value := range tests {
// 		buf := &bytes.Buffer{}
// 		enc := NewEncoder(buf)
// 		enc.WriteInt64(value)
// 		output := buf.String()
// 		// t.Logf("%q", output)
// 		total += len(output)
// 	}

// 	t.Logf("Average: %v", float64(total)/float64(len(tests)))

// 	total = 0

// 	for _, value := range tests {
// 		buf := &bytes.Buffer{}
// 		enc := gob.NewEncoder(buf)
// 		enc.Encode(value)
// 		total += len(buf.String())
// 	}

// 	t.Logf("Average: %v", float64(total)/float64(len(tests)))

// }

// func TestStringSlice(t *testing.T) {

// 	input := []string{"hello", "world", "hehe", "okay"}
// 	buf := &bytes.Buffer{}
// 	enc := NewEncoder(buf)
// 	dec := &Decoder{buf}

// 	err := enc.WriteStringSlice(input)
// 	if err != nil {
// 		t.Errorf("Got error encoding string slice: %s", err)
// 		return
// 	}

// 	result, err := dec.ReadStringSlice()
// 	if err != nil {
// 		t.Errorf("Got error decoding string slice: %s", err)
// 		return
// 	}

// 	if len(input) != len(result) {
// 		t.Errorf("Got mis-matched result for string slice: %#v -> %#v", input, result)
// 		return
// 	}

// 	for idx, item := range input {
// 		if item != result[idx] {
// 			t.Errorf("Got mis-matched result for string slice: %#v -> %#v", input, result)
// 			return
// 		}
// 	}

// }

// func TestFoo(t *testing.T) {

// 	x := map[string]interface{}{
// 		"foo": true,
// 	}

// 	buf := &bytes.Buffer{}
// 	enc := NewEncoder(buf)
// 	err := enc.WriteDict(x)

// 	if err != nil {
// 		t.Errorf("Got error encoding dict map: %s", err)
// 		return
// 	}

// 	t.Logf(string(buf.Bytes()))

// }

// func TestReadInt(t *testing.T) {

// 	tests := map[string]uint64{
// 		"\x00":                                     0,
// 		"\x95\x9a\xef:":                            123456789,
// 		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\x01": 18446744073709551615,
// 	}

// 	for value, expected := range tests {
// 		buf := bytes.NewBuffer([]byte(value))
// 		dec := &Decoder{buf}
// 		result, err := dec.ReadInt()
// 		if err != nil {
// 			t.Errorf("Got error decoding %q: %s", value, err)
// 		}
// 		if result != expected {
// 			t.Errorf("Got unexpected decoding for %q: %d", value, result)
// 		}
// 	}

// }

// func testWriteInt(t *testing.T) {
// 	N := int64(8322944)
// 	buf := Buffer()
// 	WriteInt(N, buf)
// 	fmt.Printf("%q\n", string(buf.Bytes()))
// }

// func testWriteBigInt(t *testing.T) {
// 	N := big.NewInt(8322944)
// 	buf := Buffer()
// 	WriteBigInt(N, buf)
// 	fmt.Printf("%q\n", string(buf.Bytes()))
// }

// func testWriteIntOrdering(t *testing.T) {

// 	buf := Buffer()
// 	WriteInt(-10258176, buf)
// 	prev := string(buf.Bytes())

// 	var i int64

// 	for i = -10258175; i < 10258175; i++ {
// 		buf.Reset()
// 		WriteInt(i, buf)
// 		cur := string(buf.Bytes())
// 		if prev >= cur {
// 			t.Errorf("Lexicographical ordering failure for %d -- %q >= %q", i, prev, cur)
// 		}
// 		prev = cur
// 	}

// }

// func testWriteBigIntOrdering(t *testing.T) {

// 	buf := Buffer()
// 	WriteBigInt(big.NewInt(-10258176), buf)
// 	prev := string(buf.Bytes())

// 	var i int64

// 	for i = -10258175; i < 10258175; i++ {
// 		buf.Reset()
// 		WriteBigInt(big.NewInt(i), buf)
// 		cur := string(buf.Bytes())
// 		if prev >= cur {
// 			t.Errorf("Lexicographical ordering failure for %d -- %q >= %q", i, prev, cur)
// 		}
// 		prev = cur
// 	}

// }

// func decimal(value string) *big.Decimal {
// 	dec, _ := big.NewDecimal(value)
// 	return dec
// }

// func TestWriteDecimalOrdering(t *testing.T) {

// 	buf := Buffer()
// 	WriteDecimal(decimal("0"), buf)
// 	prev := string(buf.Bytes())

// 	tests := []string{
// 		"0.02",
// 		"0.0201",
// 		"0.05",
// 		"2",
// 		"2.30001",
// 		"2.30002",
// 	}

// 	for _, value := range tests {
// 		buf.Reset()
// 		WriteDecimal(decimal(value), buf)
// 		cur := string(buf.Bytes())
// 		if prev >= cur {
// 			left, right := decimal(value).Components()
// 			t.Errorf("Lexicographical ordering failure for %s (%s, %s) -- %q >= %q",
// 				value, left, right, prev, cur)
// 		}
// 		prev = cur
// 	}

// }

// func BenchmarkWriteInt(b *testing.B) {
// 	buf := Buffer()
// 	enc := NewEncoder(buf)
// 	for i := 0; i < b.N; i++ {
// 		buf.Reset()
// 		enc.WriteInt(123456789)
// 	}
// }

// func BenchmarkWriteNumber(b *testing.B) {
// 	buf := Buffer()
// 	for i := 0; i < b.N; i++ {
// 		buf.Reset()
// 		WriteNumber("123456789", buf)
// 	}
// }

// func BenchmarkWriteNumber(b *testing.B) {
// 	// var n uint64 = 18446744073709551615
// 	buf := &bytes.Buffer{}
// 	for i := 0; i < b.N; i++ {
// 		WriteUint64(uint64(i), buf)
// 	}
// }
