// Public Domain (-) 2018-present, The Core Authors.
// See the Core UNLICENSE file for details.

package bytesize

import (
	"testing"
)

func TestInt(t *testing.T) {
	for _, tt := range []struct {
		err  bool
		v    Value
		want int
	}{
		{false, 100 * KB, 102400},
		{false, 2048 * KB, 2097152},
		{false, 32 * MB, 33554432},
		{false, 32 * GB, 34359738368},
		{false, 32 * TB, 35184372088832},
		{false, 32 * PB, 36028797018963968},
		{false, MB + 1234, 1049810},
		{true, 15360 * PB, 0},
	} {
		v, err := tt.v.Int()
		if err != nil {
			if tt.err {
				continue
			}
			t.Errorf("unexpected error when getting int value from %v: %s", tt.v, err)
			continue
		}
		if v != tt.want {
			t.Errorf("Int(%q) = %d: want %d", tt.v, v, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	for _, tt := range []struct {
		err  bool
		v    string
		want Value
	}{
		{false, "0B", 0},
		{false, "0PB", 0},
		{false, "100KB", 100 * KB},
		{false, "2MB", 2048 * KB},
		{false, "2048KB", 2048 * KB},
		{true, "2048 KB", 0},
		{true, "2048 kilobytes", 0},
		{false, "2048", 2 * KB},
		{false, "32MB", 32 * MB},
		{false, "32GB", 32 * GB},
		{false, "32TB", 32 * TB},
		{false, "8096PB", 8096 * PB},
		{true, "8096000PB", 0},
		{true, "8096 MiB", 0},
		{true, "0x1fa0 MiB", 0},
	} {
		v, err := Parse(tt.v)
		if err != nil {
			if tt.err {
				continue
			}
			t.Errorf("unexpected error when parsing %q: %s", tt.v, err)
			continue
		}
		if v != tt.want {
			t.Errorf("Parse(%q) = %q: want %q", tt.v, v, tt.want)
		}
	}
}

func TestString(t *testing.T) {
	for _, tt := range []struct {
		v    Value
		want string
	}{
		{0, "0B"},
		{0 * PB, "0B"},
		{100 * KB, "100KB"},
		{2048 * KB, "2MB"},
		{32 * MB, "32MB"},
		{32 * GB, "32GB"},
		{32 * TB, "32TB"},
		{32 * PB, "32PB"},
		{MB + 1234, "1049810B"},
	} {
		repr := tt.v.String()
		if repr != tt.want {
			t.Errorf("String(%q) = %q: want %q", tt.v, repr, tt.want)
		}
	}
}
