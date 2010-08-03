// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// Ampify Runtime
// ==============

// The runtime package provides utilities to manage the runtime environment for
// a given Ampify process/application.
package runtime

import (
	"amp/command"
	"fmt"
	"os"
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
		for _, line := range strings.Split(output, "\n", 0) {
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
		fmt.Print(
			"ERROR: The AMPIFY_ROOT environment variable hasn't been set.\n")
		os.Exit(1)
	}
}
