// Public Domain (-) 2018-present, The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// Package osexit mocks the os.Exit function.
//
// To use, first set a package-specific osExit function, e.g.
//
//     var osExit = os.Exit
//
// Then use that wherever you want to call os.Exit, e.g.
//
//     if somethingFatal {
//         osExit(1)
//         return
//     }
//
// Make sure to return immediately after the call to osExit, so that testing
// code will match real code as closely as possible.
//
// You can now use the utility functions provided by this package to override
// osExit for testing purposes, e.g.
//
//     osExit = osexit.Set()
//     invokeCodeCallingExit()
//     if !osexit.Called() {
//         t.Fatalf("os.Exit was not called as expected")
//     }
//
package osexit

import (
	"sync"
)

var (
	called bool
	mu     sync.RWMutex // protects called, status
	status int
)

// Called returns whether the mock os.Exit function was called.
func Called() bool {
	mu.RLock()
	c := called
	mu.RUnlock()
	return c
}

// Func provides a mock for the os.Exit function. Special care must be taken
// when testing os.Exit to make sure no code runs after the call to Exit. It's
// recommended to put a return statement after Exit calls so that the behaviour
// of the mock matches that of the real function as much as possible.
func Func(code int) {
	mu.Lock()
	if called {
		mu.Unlock()
		return
	}
	called = true
	status = code
	mu.Unlock()
}

// Reset resets the state of the mock function.
func Reset() {
	mu.Lock()
	called = false
	status = 0
	mu.Unlock()
}

// Set returns the mock os.Exit function after calling Reset.
func Set() func(int) {
	Reset()
	return Func
}

// Status returns the status code that the mock os.Exit function was called
// with.
func Status() int {
	mu.RLock()
	s := status
	mu.RUnlock()
	return s
}
