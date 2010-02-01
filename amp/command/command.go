// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// The command package provides utility functions for dealing with executing
// system commands
package command

import (
	"fmt"
	"io/ioutil"
	"os"
)

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
