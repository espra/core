// Public Domain (-) 2018-present, The Amp Authors.
// See the Amp UNLICENSE file for details.

package overflow

import (
	"testing"
)

func TestMulU64(t *testing.T) {
	for _, tt := range []struct {
		a    uint64
		b    uint64
		want bool
	}{
		{4294967291, 4294967271, true},
		{4294967291, 4294967321, false},
	} {
		v, ok := MulU64(tt.a, tt.b)
		if ok != tt.want {
			t.Errorf("MulU64(%d, %d) = %d: want %v", tt.a, tt.b, v, tt.want)
		}
	}
}
