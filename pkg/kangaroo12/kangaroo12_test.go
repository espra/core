// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

package kangaroo12

import (
	"encoding/hex"
	"math"
	"testing"

	"dappui.com/pkg/crypto"
)

var tests = []struct {
	custom string
	data   []byte
	length int
	want   string
}{
	{"", nil, 32, "1ac2d450fc3b4205d19da7bfca1b37513c0803577ac7167f06fe2ce1f0ef39e5"},
	{"", nil, 64, "1ac2d450fc3b4205d19da7bfca1b37513c0803577ac7167f06fe2ce1f0ef39e54269c056b8c82e48276038b6d292966cc07a3d4645272e31ff38508139eb0a71"},
	{"", nil, 10032, "e8dc563642f7228c84684c898405d3a834799158c079b12880277a1d28e2ff6d"},
	{"", genData(0), 32, "2bda92450e8b147f8a7cb629e784a058efca7cf7d8218e02d345dfaa65244a1f"},
	{"", genData(1), 32, "6bf75fa2239198db4772e36478f8e19b0f371205f6a9a93a273f51df37122888"},
	{"", genData(2), 32, "0c315ebcdedbf61426de7dcf8fb725d1e74675d7f5327a5067f367b108ecb67c"},
	{"", genData(3), 32, "cb552e2ec77d9910701d578b457ddf772c12e322e4ee7fe417f92c758f0d59d0"},
	{"", genData(4), 32, "8701045e22205345ff4dda05555cbb5c3af1a771c2b89baef37db43d9998b9fe"},
	{"", genData(5), 32, "844d610933b1b9963cbdeb5ae3b6b05cc7cbd67ceedf883eb678a0a8e0371682"},
	{"", genData(6), 32, "3c390782a8a4e89fa6367f72feaaf13255c8d95878481d3cd8ce85f58e880af8"},
	{genCustom(0), nil, 32, "fab658db63e94a246188bf7af69a133045f46ee984c56e3c3328caaf1aa1a583"},
	{genCustom(1), gen0xff(1), 32, "d848c5068ced736f4462159b9867fd4c20b808acc3d5bc48e0b06ba0a3762ec4"},
	{genCustom(2), gen0xff(3), 32, "c389e5009ae57120854c2e8c64670ac01358cf4c1baf89447a724234dc7ced74"},
	{genCustom(3), gen0xff(7), 32, "75d2f86a2e644566726b4fbcfc5657b9dbcf070c7b0dca06450ab291d7443bcf"},
}

func TestHash(t *testing.T) {
	h := crypto.Kangaroo12.New()
	size := h.BlockSize()
	if size != 168 {
		t.Errorf("unexpected block size value: got %d, want 168", size)
	}
	size = h.Size()
	if size != 32 {
		t.Errorf("unexpected digest size value: got %d, want 32", size)
	}
	const edgecase = 8184
	data := [edgecase]byte{}
	h.Write(data[:])
	digest := hex.EncodeToString(h.Sum(nil))
	want := "0ee0496ec20705da30be2e03279dae386148b9cca11eb152167ed932c23cd65f"
	if digest != want {
		t.Errorf("got mismatching digest: got %q, want %q", digest, want)
	}
	xof := h.XOF()
	if _, err := xof.ReadAt(nil, 1024); err == nil {
		t.Errorf("unexpected successful ReadAt call")
	}
	h.Reset()
	h.Write(data[:4092])
	clone := h.Clone()
	h.Reset()
	clone.Write(data[:4092])
	digest = hex.EncodeToString(clone.Sum(nil))
	if digest != want {
		t.Errorf("got mismatching digest: got %q, want %q", digest, want)
	}
}

func TestVectors(t *testing.T) {
	for i, tt := range tests {
		var h crypto.Hash
		if len(tt.custom) == 0 {
			h = New()
		} else {
			h = Custom(tt.custom)
		}
		h.Write(tt.data)
		digest := ""
		if tt.length == 32 {
			digest = hex.EncodeToString(h.Sum(nil))
		} else {
			xof := h.XOF()
			buf := make([]byte, tt.length)
			xof.Read(buf)
			if tt.length > 64 {
				buf = buf[tt.length-32:]
			}
			digest = hex.EncodeToString(buf)
		}
		if digest != tt.want {
			t.Errorf("got mismatching digest for test vector %d: got %q, want %q", i+1, digest, tt.want)
		}
	}
}

func TestVectorsParallel(t *testing.T) {
	for i, tt := range tests {
		xof := Parallel(4, []byte(tt.custom), tt.data)
		buf := make([]byte, tt.length)
		xof.Read(buf)
		if tt.length > 64 {
			buf = buf[tt.length-32:]
		}
		digest := hex.EncodeToString(buf)
		if digest != tt.want {
			t.Errorf("got mismatching digest for test vector %d: got %q, want %q", i+1, digest, tt.want)
		}
	}
}

func gen0xff(l int) []byte {
	out := make([]byte, l)
	for i := 0; i < l; i++ {
		out[i] = 0xff
	}
	return out
}

func genCustom(exp float64) string {
	return string(genPattern(41, exp))
}

func genData(exp float64) []byte {
	return genPattern(17, exp)
}

func genPattern(base float64, exp float64) []byte {
	l := int(math.Pow(base, exp))
	out := make([]byte, l)
	char := byte(0)
	for i := 0; i < l; i++ {
		out[i] = char
		if char == 0xfa {
			char = 0
		} else {
			char++
		}
	}
	return out
}
