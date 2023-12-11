// Public Domain (-) 2010-present, The Espra Core Authors.
// See the Espra Core UNLICENSE file for details.

package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"

	"espra.dev/pkg/osexit"
)

func TestCrash(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("Crash didn't generate an abort error")
		}
	}()
	Crash()
}

func TestCreatePIDFile(t *testing.T) {
	reset()
	dir := mktemp(t)
	defer os.RemoveAll(dir)
	fpath := filepath.Join(dir, "test.pid")
	err := CreatePIDFile(fpath)
	if err != nil {
		t.Fatalf("Unexpected error creating PID file: %s", err)
	}
	written, err := os.ReadFile(fpath)
	if err != nil {
		t.Fatalf("Unexpected error reading PID file: %s", err)
	}
	expected := os.Getpid()
	pid, err := strconv.ParseInt(string(written), 10, 64)
	if err != nil {
		t.Fatalf("Unexpected error parsing PID file contents as an int: %s", err)
	}
	if int(pid) != expected {
		t.Fatalf("Mismatching PID file contents: got %d, want %d", int(pid), expected)
	}
	Exit(2)
	if !osexit.Called() || osexit.Status() != 2 {
		t.Fatalf("Exit call did not behave as expected")
	}
	_, err = os.Stat(fpath)
	if err == nil {
		t.Fatalf("Calling Exit did not remove the created PID file as expected")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("Calling Exit did not remove the created PID file as expected, got error: %s", err)
	}
	fpath = filepath.Join(dir+"-nonexistent-directory", "test.pid")
	err = CreatePIDFile(fpath)
	if err == nil {
		t.Fatalf("Expected an error when creating PID file in a non-existent directory")
	}
}

func TestDisableDefaultExit(t *testing.T) {
	reset()
	called := false
	SetExitHandler(func() {
		called = true
	})
	send(syscall.SIGTERM)
	if !osexit.Called() {
		t.Fatalf("os.Exit was not called on SIGTERM")
	}
	if !called {
		t.Fatalf("Exit handler not run on SIGTERM")
	}
	DisableAutoExit()
	osexit.Reset()
	called = false
	resetExiting()
	send(syscall.SIGTERM)
	if osexit.Called() {
		t.Fatalf("os.Exit was called on SIGTERM even after DisableAutoExit()")
	}
	if !called {
		t.Fatalf("Exit handler not run on SIGTERM after DisableAutoExit")
	}
}

func TestExit(t *testing.T) {
	reset()
	called := false
	SetExitHandler(func() {
		called = true
	})
	Exit(7)
	if !osexit.Called() {
		t.Fatalf("Exit did not call os.Exit")
	}
	status := osexit.Status()
	if status != 7 {
		t.Fatalf("Exit did not set the right status code: got %d, want 7", status)
	}
	if !called {
		t.Fatalf("Exit handler was not run when calling Exit")
	}
	osexit.Reset()
	called = false
	go func() {
		Exit(8)
	}()
	<-testSig
	wait <- struct{}{}
	if osexit.Called() {
		t.Fatalf("Second call to Exit called os.Exit")
	}
	if called {
		t.Fatalf("Second call to Exit resulted in Exit handler being run again")
	}
}

func TestInit(t *testing.T) {
	dir := mktemp(t)
	defer os.RemoveAll(dir)
	err := Init(dir, "core")
	if err != nil {
		t.Fatalf("Unexpected error initialising process: %s", err)
	}
	err = Init(dir+"-nonexistent-directory", "core")
	if err == nil {
		t.Fatalf("Expected an error when calling Init in a non-existing directory")
	}
}

func TestLock(t *testing.T) {
	reset()
	dir := mktemp(t)
	defer os.RemoveAll(dir)
	err := Lock(dir, "core")
	if err != nil {
		t.Fatalf("Unexpected error acquiring Lock: %s", err)
	}
	err = Lock(dir, "core")
	if err == nil {
		t.Fatalf("Expected an error when calling Lock on an already locked path")
	}
	fpath := filepath.Join(dir, fmt.Sprintf("core-%d.lock", os.Getpid()))
	_, err = os.Stat(fpath)
	if err != nil {
		t.Fatalf("Unexpected error accessing the raw lock file: %s", err)
	}
	Exit(2)
	_, err = os.Stat(fpath)
	if err == nil {
		t.Fatalf("Calling Exit did not remove the lock file as expected")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("Calling Exit did not remove the lock file as expected, got error: %s", err)
	}
	err = Lock(dir+"-nonexistent-directory", "core")
	if err == nil {
		t.Fatalf("Expected an error when calling Lock in a non-existing directory")
	}
}

func TestSignalHandler(t *testing.T) {
	reset()
	called := false
	SetSignalHandler(syscall.SIGHUP, func() {
		called = true
	})
	send(syscall.SIGABRT)
	if called {
		t.Fatalf("Signal handler erroneously called on SIGABRT")
	}
	send(syscall.SIGHUP)
	if !called {
		t.Fatalf("Signal handler not called on SIGHUP")
	}
}

func mktemp(t *testing.T) string {
	dir, err := os.MkdirTemp("", "core-process")
	if err != nil {
		t.Skipf("Unable to create temporary directory for tests: %s", err)
	}
	return dir
}

func reset() {
	OSExit = osexit.Set()
	testMode = true
	ResetHandlers()
	resetExiting()
}

func resetExiting() {
	mu.Lock()
	exiting = false
	mu.Unlock()
}

func send(sig syscall.Signal) {
	syscall.Kill(syscall.Getpid(), sig)
	<-testSig
}
