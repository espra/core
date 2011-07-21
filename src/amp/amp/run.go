// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package main

import (
	"amp/optparse"
	"amp/runtime"
	"exec"
	"path/filepath"
	"os"
	"time"
)

func runProcess(amp, cmd, path, config string, console bool) {

	var files []*os.File

	if console {
		files = []*os.File{nil, os.Stdout, os.Stderr}
	} else {
		files = []*os.File{nil, nil, nil}
	}

	process, err := os.StartProcess(
		amp,
		[]string{"amp", cmd, config},
		&os.ProcAttr{
			Dir:   path,
			Env:   os.Environ(),
			Files: files,
		})

	if err != nil {
		runtime.StandardError(err)
	}

	_, err = process.Wait(0)
	if err != nil {
		runtime.StandardError(err)
	}
}

func run(argv []string, usage string) {

	opts := optparse.Parser(
		"Usage: amp run <instance-path> [options]\n\n    " + usage + "\n")

	profile := opts.String([]string{"--profile"}, "development",
		"the config profile to use [development]", "NAME")

	frontendPath := opts.String([]string{"--frontend"}, "frontend",
		"the path to the frontend directory [frontend]", "PATH")

	nodePath := opts.String([]string{"--node"}, "node",
		"the path to the node directory [node]", "PATH")

	noConsoleLog := opts.Bool([]string{"--no-console-log"}, false,
		"disable output to stdout/stderr")

	args := opts.Parse(argv)

	if len(args) == 0 || args[0] == "help" {
		opts.PrintUsage()
		runtime.Exit(0)
	}

	root, err := filepath.Abs(filepath.Clean(args[0]))
	if err != nil {
		runtime.StandardError(err)
	}

	_, err = os.Open(root)
	if err != nil {
		runtime.StandardError(err)
	}

	frontendDirectory := runtime.JoinPath(root, *frontendPath)
	_, err = os.Open(frontendDirectory)
	if err != nil {
		runtime.StandardError(err)
	}

	nodeDirectory := runtime.JoinPath(root, *nodePath)
	_, err = os.Open(nodeDirectory)
	if err != nil {
		runtime.StandardError(err)
	}

	amp, err := exec.LookPath("amp")
	if err != nil {
		runtime.StandardError(err)
	}

	runtime.Init()
	config := *profile + ".yaml"

	console := true
	if *noConsoleLog {
		console = false
	}

	go runProcess(amp, "frontend", frontendDirectory, config, console)

	<-time.After(2000000000)

	go runProcess(amp, "node", nodeDirectory, config, console)

	// Enter the wait loop for the process to be killed.
	loopForever := make(chan bool, 1)
	<-loopForever

}
