// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package main

import (
	"amp/logging"
	"amp/master"
	"amp/nodule"
	"amp/optparse"
	"amp/runtime"
	"os"
	"strings"
)

func ampNode(argv []string, usage string) {

	opts := optparse.Parser(
		"Usage: amp node <config.yaml> [options]\n\n    " + usage + "\n")

	nodeHost := opts.StringConfig("node-host", "",
		"the host to bind this node to")

	nodePort := opts.IntConfig("node-port", 8050,
		"the port to bind this node to [8050]")

	ctrlHost := opts.StringConfig("control-host", "",
		"the host to bind the nodule control socket to")

	ctrlPort := opts.IntConfig("control-port", 8051,
		"the port to bind the nodule control socket to [8051]")

	nodules := opts.StringConfig("nodules", "*",
		"comma-separated list of nodules to initialise [*]")

	nodulePaths := opts.StringConfig("nodule-paths", ".",
		"comma-separated list of nodule container directories [nodule]")

	masterNodes := opts.StringConfig("master-nodes", "localhost:8060",
		"comma-separated addresses of amp master nodes [localhost:8060]")

	masterKeyPath := opts.StringConfig("master-key", "cert/master.key",
		"the path to the file containing the amp master public key [cert/master.key]")

	debug, _, runPath := runtime.DefaultOpts("node", opts, argv, nil)

	masterClient, err := master.NewClient(*masterNodes, *masterKeyPath)

	if err != nil {
		runtime.StandardError(err)
	}

	logging.AddConsoleFilter(nodule.FilterConsoleLog)

	node, err := nodule.NewHost(
		runPath, *nodeHost, *nodePort, *ctrlHost, *ctrlPort, *nodules,
		strings.SplitN(*nodulePaths, ",", -1), masterClient)

	if err != nil {
		runtime.StandardError(err)
	}

	err = node.Run(debug)
	logging.Wait()

	if err != nil {
		runtime.Exit(1)
	}

}

func getNodules(paths []string) (nodules []*nodule.Nodule) {

	if paths == nil {
		cwd, err := os.Getwd()
		if err != nil {
			runtime.StandardError(err)
		}
		paths = []string{cwd}
	}

	nodules = make([]*nodule.Nodule, 0)
	seen := make(map[string]bool)

	for _, path := range paths {
		data, err := nodule.Find(path)
		if err != nil {
			runtime.StandardError(err)
		}
		for _, nodule := range data {
			if seen[nodule.Path] {
				continue
			}
			nodules = append(nodules, nodule)
			seen[nodule.Path] = true
		}
	}

	return

}

func handleCommon(name, usage string, argv []string) (args []string) {

	opts := optparse.Parser(
		"Usage: amp " + name + " <path> [options]\n\n    " + usage + "\n")

	profile := opts.String([]string{"--profile"}, "development",
		"the config profile to use [development]", "NAME")

	noConsoleLog := opts.Bool([]string{"--no-console-log"}, false,
		"disable output to stdout/stderr")

	args = opts.Parse(argv)

	if len(args) == 0 {
		opts.PrintUsage()
		runtime.Exit(0)
	}

	runtime.SetProfile(*profile)

	if !*noConsoleLog {
		logging.AddConsoleLogger()
		logging.AddConsoleFilter(nodule.FilterConsoleLog)
	}

	return

}

func ampBuild(argv []string, usage string) {

	args := handleCommon("build", usage, argv)
	status := 0

	for _, nodule := range getNodules(args) {
		err := nodule.Build()
		if err != nil {
			status = 1
			break
		}
	}

	logging.Wait()
	runtime.Exit(status)

}

func ampTest(argv []string, usage string) {

	args := handleCommon("test", usage, argv)
	status := 0

	for _, nodule := range getNodules(args) {
		err := nodule.Test()
		if err != nil {
			status = 1
			break
		}
	}

	logging.Wait()
	runtime.Exit(status)

}

func ampReview(argv []string, usage string) {

	args := handleCommon("review", usage, argv)
	status := 0

	for _, nodule := range getNodules(args) {
		err := nodule.Review()
		if err != nil {
			status = 1
			break
		}
	}

	logging.Wait()
	runtime.Exit(status)

}
