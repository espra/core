// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// The runtime package provides utilities to manage the runtime environment for
// a given Go process/application.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
)

var Platform = os.Getenv("GOOS")

type CommandError struct {
	Command string
	Args []string
}

func (err *CommandError) String() string {
	return fmt.Sprintf("Couldn't successfully execute: %s %v", err.Command, err.Args)
}

// Return the output from running the given command
func GetOutput(cmd string, args []string) (output []byte, error os.Error) {
	read_pipe, write_pipe, err := os.Pipe()
	if err != nil {
		goto Error
	}
	pid, err := os.ForkExec(cmd, args, os.Environ(), "", []*os.File{nil, write_pipe, nil})
	if err != nil {
		goto Error
	}
	_, err = os.Wait(pid, 0)
	if err != nil {
		goto Error
	}
	write_pipe.Close()
	output, err = ioutil.ReadAll(read_pipe)
	if err != nil {
		goto Error
	}
	read_pipe.Close()
	return output, nil
Error:
	return nil, &CommandError{cmd, args}
}

// Try and figure out the number of CPUs on the current machine
func GetCPUCount() (count int) {
	if (Platform == "darwin") || (Platform == "freebsd") {
		output, err := GetOutput("/usr/sbin/sysctl", []string{"sysctl", "-n", "hw.ncpu"})
		if err == nil {
			output = bytes.TrimSpace(output)
			if _cpus, err := strconv.Atoi(string(output)); err == nil {
				count = _cpus
			}
		}
	} else if Platform == "linux" {
		output, err := GetOutput("/bin/cat", []string{"cat", "/proc/cpuinfo"})
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

func init() {
	runtime.GOMAXPROCS(GetCPUCount())
}

func main() {
	fmt.Println(GetCPUCount())
}
