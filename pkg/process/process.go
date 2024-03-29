// Public Domain (-) 2010-present, The Espra Core Authors.
// See the Espra Core UNLICENSE file for details.

// Package process provides utilities for managing the current system process.
package process

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"slices"
	"sync"
	"syscall"
)

// OSExit is the function used to terminate the current process. It defaults to
// os.Exit, but can be overridden for testing purposes.
var OSExit = os.Exit

var (
	exitDisabled bool
	exiting      bool
	handlerID    int
	mu           sync.RWMutex // protects exitDisabled, exiting, handlerID, registry
	registry     = map[os.Signal][]entry{}
	testMode     = false
	testSig      = make(chan struct{}, 10)
	wait         = make(chan struct{})
)

type entry struct {
	handler func()
	id      int
}

type lockFile struct {
	file string
	link string
}

func (l *lockFile) release() {
	os.Remove(l.file)
	os.Remove(l.link)
}

// RemoveHandler defines a function for removing a registered signal handler.
type RemoveHandler func()

// Crash will terminate the process with a panic that will generate stacktraces
// for all user-generated goroutines.
func Crash() {
	debug.SetTraceback("all")
	panic("abort")
}

// CreatePIDFile writes the current process ID to a new file at the given path.
// The written file is removed when Exit is called, or when the process receives
// an os.Interrupt or SIGTERM signal.
func CreatePIDFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o660)
	if err != nil {
		return err
	}
	fmt.Fprintf(f, "%d", os.Getpid())
	err = f.Close()
	if err == nil {
		SetExitHandler(func() {
			os.Remove(path)
		})
	}
	return err
}

// DisableAutoExit will prevent the process from automatically exiting after
// processing os.Interrupt or SIGTERM signals. This will not be enforced if Exit
// is called directly.
func DisableAutoExit() {
	mu.Lock()
	exitDisabled = true
	mu.Unlock()
}

// Exit runs the registered exit handlers, as if the os.Interrupt signal had
// been sent, and then terminates the process with the given status code. Exit
// blocks until the process terminates if it has already been called elsewhere.
func Exit(code int) {
	mu.Lock()
	if exiting {
		mu.Unlock()
		if testMode {
			testSig <- struct{}{}
		}
		<-wait
		return
	}
	exiting = true
	entries := slices.Clone(registry[os.Interrupt])
	mu.Unlock()
	for _, entry := range entries {
		entry.handler()
	}
	OSExit(code)
}

// Init tries to acquire a process lock and write the PID file for the current
// process.
func Init(directory string, name string) error {
	if err := Lock(directory, name); err != nil {
		return err
	}
	return CreatePIDFile(filepath.Join(directory, name+".pid"))
}

// Lock tries to acquire a process lock in the given directory. The acquired
// lock file is released when Exit is called, or when the process receives an
// os.Interrupt or SIGTERM signal.
//
// This function has only been tested for correctness on Unix systems with
// filesystems where link is atomic. It may not work as expected on NFS mounts
// or on platforms like Windows.
func Lock(directory string, name string) error {
	file := filepath.Join(directory, fmt.Sprintf("%s-%d.lock", name, os.Getpid()))
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0o660)
	if err != nil {
		return err
	}
	f.Close()
	link := filepath.Join(directory, name+".lock")
	err = os.Link(file, link)
	if err != nil {
		// We don't remove the lock file here so that calling Lock multiple
		// times from the same process doesn't remove an existing lock.
		return err
	}
	l := &lockFile{
		file: file,
		link: link,
	}
	SetExitHandler(l.release)
	return nil
}

// ReapOrphans reaps orphaned child processes and returns whether there are any
// unterminated child processes that are still active.
//
// This is currently a no-op on all platforms except Linux.
func ReapOrphans() bool {
	return reap()
}

// ResetHandlers drops all currently registered handlers.
func ResetHandlers() {
	mu.Lock()
	registry = map[os.Signal][]entry{}
	mu.Unlock()
}

// RunReaper continuously attempts to reap orphaned child processes until the
// given context is cancelled.
//
// On Linux, this will register the current process as a child subreaper, and
// attempt to reap child processes whenever SIGCHLD is received. On all other
// platforms, this is currently a no-op.
func RunReaper(ctx context.Context) {
	runReaper(ctx)
}

// SetExitHandler registers the given handler function to run when receiving
// os.Interrupt or SIGTERM signals. Registered handlers are executed in reverse
// order of when they were set.
func SetExitHandler(handler func()) RemoveHandler {
	mu.Lock()
	e := entry{handler, handlerID}
	handlerID++
	registry[os.Interrupt] = slices.Insert(registry[os.Interrupt], 0, e)
	registry[syscall.SIGTERM] = slices.Insert(registry[syscall.SIGTERM], 0, e)
	mu.Unlock()
	return func() {
		removeHandler(e.id, os.Interrupt, syscall.SIGTERM)
	}
}

// SetSignalHandler registers the given handler function to run when receiving
// the specified signal. Registered handlers are executed in reverse order of
// when they were set.
func SetSignalHandler(signal os.Signal, handler func()) RemoveHandler {
	mu.Lock()
	e := entry{handler, handlerID}
	handlerID++
	registry[signal] = slices.Insert(registry[signal], 0, e)
	mu.Unlock()
	return func() {
		removeHandler(e.id, signal)
	}
}

func handleSignals() {
	notifier := make(chan os.Signal, 100)
	signal.Notify(notifier)
	go func() {
		for sig := range notifier {
			mu.Lock()
			disabled := exitDisabled
			if !disabled {
				if sig == syscall.SIGTERM || sig == os.Interrupt {
					exiting = true
				}
			}
			entries := slices.Clone(registry[sig])
			mu.Unlock()
			for _, entry := range entries {
				entry.handler()
			}
			if !disabled {
				if sig == syscall.SIGTERM || sig == os.Interrupt {
					OSExit(1)
				}
			}
			if testMode {
				testSig <- struct{}{}
			}
		}
	}()
}

func removeHandler(id int, signals ...os.Signal) {
	mu.Lock()
	for _, signal := range signals {
		entries := registry[signal]
		idx := -1
		for i, entry := range entries {
			if entry.id == id {
				idx = i
				break
			}
		}
		registry[signal] = append(entries[:idx], entries[idx+1:]...)
	}
	mu.Unlock()
}

func init() {
	handleSignals()
}
