// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// Ampify Runtime
// ==============

// The runtime package provides utilities to manage the runtime environment for
// a given Ampify process/application.
package runtime

import (
	"amp/command"
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

const Platform = syscall.OS

var (
	AmpifyRoot string
	CPUCount   int
)

var signalHandlers = make(map[os.UnixSignal]func())
var exitHandlers = []func(){}

func RegisterSignalHandler(signal os.UnixSignal, handler func()) {
	signalHandlers[signal] = handler
}

func ClearSignalHandler(signal os.UnixSignal) {
	signalHandlers[signal] = func() {}, false
}

func handleSignals() {
	var sig os.Signal
	for {
		sig = <-signal.Incoming
		handler, found := signalHandlers[sig.(os.UnixSignal)]
		if found {
			handler()
		}
	}
}

func RunExitHandlers() {
	for _, handler := range exitHandlers {
		handler()
	}
}

func RegisterExitHandler(handler func()) {
	exitHandlers = append(exitHandlers, handler)
}

func Exit(code int) {
	RunExitHandlers()
	os.Exit(code)
}

// Utility function which calls Exit and matches the signal handler interface.
func exitProcess() {
	Exit(0)
}

func Error(message string, v ...interface{}) {
	if len(v) == 0 {
		fmt.Fprint(os.Stderr, message)
	} else {
		fmt.Fprintf(os.Stderr, message, v...)
	}
	Exit(1)
}

func StandardError(err os.Error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
	Exit(1)
}

func CreatePidFile(path string) {
	pidFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		Error("ERROR: %s\n", err)
	}
	fmt.Fprintf(pidFile, "%d", os.Getpid())
	err = pidFile.Close()
	if err != nil {
		Error("ERROR: %s\n", err)
	}
}

type Lock struct {
	link     string
	file     string
	acquired bool
}

func GetLock(directory string, name string) (lock *Lock, err os.Error) {
	file := path.Join(directory, fmt.Sprintf("%s-%d.lock", name, os.Getpid()))
	lockFile, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return lock, err
	}
	lockFile.Close()
	link := path.Join(directory, name+".lock")
	err = os.Link(file, link)
	if err == nil {
		lock = &Lock{
			link:     link,
			file:     file,
			acquired: true,
		}
		RegisterExitHandler(func() { lock.ReleaseLock() })
	} else {
		os.Remove(file)
	}
	return lock, err
}

func (lock *Lock) ReleaseLock() {
	if lock.acquired {
		os.Remove(lock.file)
		os.Remove(lock.link)
	}
}

// The ``JoinPath`` utility function joins the given ``path`` with the
// ``directory`` unless it happens to be an absolute path, in which case it
// returns the path exactly as it was given.
func JoinPath(directory, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(directory, filepath.Clean(path))
}

// The ``InitProcess`` utility function acquires a process lock and writes the
// PID file for the current process.
func InitProcess(name, runPath string) {

	pidFile := fmt.Sprintf("%s.pid", name)

	// Get the runtime lock to ensure we only have one process of any given name
	// running within the same run path at any time.
	_, err := GetLock(runPath, name)
	if err != nil {
		Error("ERROR: Couldn't successfully acquire a process lock:\n\n\t%s\n\n", err)
	}

	// Write the process ID into a file for use by external scripts.
	go CreatePidFile(filepath.Join(runPath, pidFile))

}


// The ``runtime.GetCPUCount`` function tries to detect the number of CPUs on
// the current machine.
func GetCPUCount() (count int) {
	// On BSD systems, it should be possible to use ``sysctl -n hw.ncpu`` to
	// figure this out.
	if (Platform == "darwin") || (Platform == "freebsd") {
		output, err := command.GetOutput(
			[]string{"/usr/sbin/sysctl", "-n", "hw.ncpu"},
		)
		if err != nil {
			return 1
		}
		count, err = strconv.Atoi(strings.TrimSpace(output))
		if err != nil {
			return 1
		}
		// Linux systems provide introspection via ``/proc/cpuinfo``.
	} else if Platform == "linux" {
		output, err := command.GetOutput([]string{"/bin/cat", "/proc/cpuinfo"})
		if err != nil {
			return 1
		}
		for _, line := range strings.Split(output, "\n", -1) {
			if strings.HasPrefix(line, "processor") {
				count += 1
			}
		}
	}
	// For unknown platforms, we assume that there's just a single processor.
	if count == 0 {
		return 1
	}
	return count
}

// A utility ``runtime.Init`` function is provided which will set Go's internal
// ``GOMAXPROCS`` to the number of CPUs detected and exit with an error message
// if the ``$AMPIFY_ROOT`` environment variable hasn't been set.
func Init() {
	runtime.GOMAXPROCS(CPUCount)
}

// -----------------------------------------------------------------------------
// Package Initialiser
// -----------------------------------------------------------------------------

// Set the ``runtime.CPUCount`` variable to the number of CPUs detected.
func init() {
	CPUCount = GetCPUCount()
	AmpifyRoot = os.Getenv("AMPIFY_ROOT")
	if AmpifyRoot == "" {
		Error("ERROR: The AMPIFY_ROOT environment variable hasn't been set.\n")
	}
	RegisterSignalHandler(os.SIGINT, exitProcess)
	RegisterSignalHandler(os.SIGTERM, exitProcess)
	go handleSignals()
}
