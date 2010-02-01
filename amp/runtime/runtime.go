// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

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

// Try and figure out the number of CPUs on the current machine
func GetCPUCount() (count int) {
	if (Platform == "darwin") || (Platform == "freebsd") {
		output, err := command.GetOutput([]string{"/usr/sbin/sysctl", "-n", "hw.ncpu"})
		if err != nil {
			return 1
		}
		count, err = strconv.Atoi(strings.TrimSpace(output))
		if err != nil {
			return 1
		}
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
	if count == 0 {
		return 1
	}
	return count
}

func Init() { runtime.GOMAXPROCS(GetCPUCount()) }
