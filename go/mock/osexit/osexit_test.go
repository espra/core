// Public Domain (-) 2018-present, The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package osexit_test

import (
	"os"
	"testing"

	"ampify.dev/go/mock/osexit"
)

var osExit = os.Exit

func TestOsExit(t *testing.T) {
	osExit = osexit.Set()
	osExit(2)
	if !osexit.Called() {
		t.Fatalf("mock exit function was not called")
	}
	status := osexit.Status()
	if status != 2 {
		t.Fatalf("mock exit function did not set the right status code: got %d, want 2", status)
	}
	osExit(3)
	status = osexit.Status()
	if status != 2 {
		t.Fatalf("mock exit function overrode the status set by a previous call: got %d, want 2", status)
	}
	osexit.Reset()
	if osexit.Called() {
		t.Fatalf("the reset mock exit function claims to have been called")
	}
	status = osexit.Status()
	if status != 0 {
		t.Fatalf("the reset mock exit function returned a non-zero status code: got %d, want 0", status)
	}
}
