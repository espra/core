// Public Domain (-) 2021-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

// +build linux

package wsl

import (
	"errors"
	"os"
	"testing"
)

func TestDetect(t *testing.T) {
	var fail error
	rel := ""
	readFile = func(path string) ([]byte, error) {
		return []byte(rel), fail
	}
	wsl := Detect()
	if wsl {
		t.Error("Detect() = true: want false")
	}
	rel = "4.19.112-microsoft-WSL2-standard"
	wsl = Detect()
	if !wsl {
		t.Errorf("Detect() = false: want true for %q", rel)
	}
	rel = "4.19.112-microsoft-standard"
	wsl = Detect()
	if !wsl {
		t.Errorf("Detect() = false: want true for %q", rel)
	}
	fail = errors.New("fail")
	wsl = Detect()
	if wsl {
		t.Errorf("Detect() = false: want true for read failure")
	}
	readFile = os.ReadFile
}
