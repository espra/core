// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package refmap

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestRefmap(t *testing.T) {

	refmap := New()
	ref1 := refmap.Incref("some value", 1)
	ref2 := refmap.Incref("some value", 3)

	if ref1 != ref2 {
		t.Errorf("Mis-matched refs for the same string: (%v, %v)", ref1, ref2)
	}

	ref3 := refmap.Incref("another value", 2)
	if ref1 == ref3 {
		t.Errorf("Got the same refs for different strings: %v", ref1)
	}

	refmap.Decref("another value", 2)
	ref4 := refmap.Incref("another value", 2)
	if ref3 == ref4 {
		t.Errorf("Got the same refs for a fully decref-ed string: %v", ref3)
	}

	ref5 := refmap.Get("another value")
	if ref4 != ref5 {
		t.Errorf("Mis-matched refs for Incref() and Get(): (%v, %v)", ref4, ref5)
	}

}

func TestPerf(t *testing.T) {
	runtime.GOMAXPROCS(4)
	N := 10000
	refmap := New()
	start := time.Nanoseconds()
	results := make(chan uint64, N)
	for i := 0; i < N; i++ {
		go func() {
			results <- refmap.Incref("some string", 2)
			refmap.Get("some string")
			refmap.Get("another string")
			refmap.Decref("some string", 1)
		}()
	}
	for i := 0; i < N; i++ {
		<-results
	}
	fmt.Printf("Took: %v\n", time.Nanoseconds()-start)
	runtime.GOMAXPROCS(1)
}
