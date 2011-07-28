// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package main

import (
	"amp/logging"
	"amp/optparse"
	"amp/runtime"
	"exec"
	"path/filepath"
	"os"
	"time"
)

func runProcess(amp, cmd, path, config string, console bool, quit chan bool) {

	var files []*os.File

	if console {
		files = []*os.File{nil, os.Stdout, os.Stderr}
	} else {
		files = []*os.File{nil, nil, nil}
	}

	logging.Info("Running: amp %s %s", cmd, config)

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

	waitmsg, err := process.Wait(0)
	if err != nil {
		runtime.StandardError(err)
	}

	if waitmsg.ExitStatus() != 0 {
		runtime.Error("ERROR: Got %s when running `amp %s %s`\n",
			waitmsg, cmd, config)
	}

	quit <- true

}

func ensureDirectory(root, path string) (directory string) {
	if path == "" {
		directory = root
	} else {
		directory = runtime.JoinPath(root, path)
	}
	_, err := os.Open(directory)
	if err != nil {
		runtime.StandardError(err)
	}
	return
}

func ampRun(argv []string, usage string) {

	opts := optparse.Parser(
		"Usage: amp run <instance-path> [options]\n\n    " + usage + "\n")

	profile := opts.String([]string{"--profile"}, "development",
		"the config profile to use [development]", "NAME")

	masterPath := opts.String([]string{"--master"}, "master",
		"the path to the amp master node directory [master]", "PATH")

	nodePath := opts.String([]string{"--node"}, "node",
		"the path to the amp node directory [node]", "PATH")

	frontendPath := opts.String([]string{"--frontend"}, "frontend",
		"the path to the amp frontend directory [frontend]", "PATH")

	noConsoleLog := opts.Bool([]string{"--no-console-log"}, false,
		"disable output to stdout/stderr")

	args := opts.Parse(argv)

	if len(args) == 0 {
		opts.PrintUsage()
		runtime.Exit(0)
	}

	root, err := filepath.Abs(filepath.Clean(args[0]))
	if err != nil {
		runtime.StandardError(err)
	}

	amp, err := exec.LookPath("amp")
	if err != nil {
		runtime.StandardError(err)
	}

	config := *profile + ".yaml"
	quit := make(chan bool, 1)

	var console bool

	if !*noConsoleLog {
		logging.AddConsoleLogger()
		console = true
	}

	runtime.Init()
	ensureDirectory(root, "")

	for _, spec := range [][]string{
		{"master", ensureDirectory(root, *masterPath)},
		{"node", ensureDirectory(root, *nodePath)},
		{"frontend", ensureDirectory(root, *frontendPath)},
	} {
		go runProcess(amp, spec[0], spec[1], config, console, quit)
		<-time.After(2000000000)
	}

	// Enter the wait loop for the process to be killed.
	<-quit

}
