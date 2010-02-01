// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// The runtime package provides utilities to manage the runtime environment for
// a given Go process/application.
package runtime

import (
	"amp/command"
	"bytes"
	"runtime"
	"strconv"
	"syscall"
)

const Platform = syscall.OS

// Try and figure out the number of CPUs on the current machine
func GetCPUCount() (count int) {
	if (Platform == "darwin") || (Platform == "freebsd") {
		output, err := command.GetOutput([]string{"/usr/sbin/sysctl", "-n", "hw.ncpu"})
		if err == nil {
			output = bytes.TrimSpace(output)
			if _cpus, err := strconv.Atoi(string(output)); err == nil {
				count = _cpus
			}
		}
	} else if Platform == "linux" {
		output, err := command.GetOutput([]string{"/bin/cat", "/proc/cpuinfo"})
		if err == nil {
			split_output := bytes.Split(output, []byte{'\n'}, 0)
			for _, line := range split_output {
				if bytes.HasPrefix(line, []byte{'p', 'r', 'o', 'c', 'e', 's', 's', 'o', 'r'}) {
					count += 1
				}
			}
		}
	}
	if count == 0 {
		return 1
	}
	return count
}

func Init() {
	runtime.GOMAXPROCS(GetCPUCount())
}
