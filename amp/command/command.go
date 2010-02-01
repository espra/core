// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// The command package provides utility functions for dealing with executing
// system commands
package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type CommandError struct {
	Command string
	Args    []string
}

func (err *CommandError) String() string {
	return fmt.Sprintf("Couldn't successfully execute: %s %v", err.Command, err.Args)
}

// Return the output from running the given command
func GetOutput(args []string) (output string, error os.Error) {
	read_pipe, write_pipe, err := os.Pipe()
	if err != nil {
		goto Error
	}
	defer read_pipe.Close()
	pid, err := os.ForkExec(args[0], args, os.Environ(), ".", []*os.File{nil, write_pipe, nil})
	if err != nil {
		write_pipe.Close()
		goto Error
	}
	_, err = os.Wait(pid, 0)
	write_pipe.Close()
	if err != nil {
		goto Error
	}
	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, read_pipe)
	if err != nil {
		goto Error
	}
	output = buffer.String()
	return output, nil
Error:
	return "", &CommandError{args[0], args}
}
