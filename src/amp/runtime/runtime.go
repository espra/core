// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// Ampify Runtime
// --------------

// The runtime package provides utilities to manage the runtime environment for
// a given Go process/application.
package runtime

import (
	"amp/command"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

const Platform = syscall.OS

// Try and figure out the number of CPUs on the current machine.
func GetCPUCount() (count int) {
	// On BSD systems, it should be possible to use ``sysctl -n hw.ncpu`` to
	// figure this out.
	if (Platform == "darwin") || (Platform == "freebsd") {
		output, err := command.GetOutput(
			[]string{"/usr/sbin/sysctl", "-n", "hw.ncpu"}
		)
		if err != nil {
			return 1
		}
		count, err = strconv.Atoi(strings.TrimSpace(output))
		if err != nil {
			return 1
		}
	// Linux provides introspection via ``/proc/cpuinfo``.
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
	// If on an unknown platform, we assume that there's just a single
	// processor.
	if count == 0 {
		return 1
	}
	return count
}

// A utility function ``runtime.Init`` is provided which will set Go's internal
// ``GOMAXPROCS`` to the number of CPUs detected.
func Init() { runtime.GOMAXPROCS(GetCPUCount()) }
