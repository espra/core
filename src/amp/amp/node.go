// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package main

import (
	"amp/nodule"
	"amp/optparse"
	"amp/repo"
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

	repoNodes := opts.StringConfig("repo-nodes", "localhost:8060",
		"comma-separated addresses of amp repo nodes [localhost:8060]")

	repoKeyPath := opts.StringConfig("repo-key", "cert/repo.key",
		"the path to the file containing the amp repo public key [cert/repo.key]")

	debug, _, runPath := runtime.DefaultOpts("node", opts, argv, nil)

	repoClient, err := repo.NewClient(*repoNodes, *repoKeyPath)

	if err != nil {
		runtime.StandardError(err)
	}

	node, err := nodule.NewHost(
		runPath, *nodeHost, *nodePort, *ctrlHost, *ctrlPort, *nodules,
		strings.SplitN(*nodulePaths, ",", -1), repoClient)

	if err != nil {
		runtime.StandardError(err)
	}

	node.Run(debug)

}

func getNodules(paths []string) (nodules [][]string) {

	if paths == nil {
		cwd, err := os.Getwd()
		if err != nil {
			runtime.StandardError(err)
		}
		paths = []string{cwd}
	}

	nodules = make([][]string, 0)
	seen := make(map[string]bool)

	for _, path := range paths {
		data, err := nodule.Find(path)
		if err != nil {
			runtime.StandardError(err)
		}
		for name, confpath := range data {
			if seen[confpath] {
				continue
			}
			nodules = append(nodules, []string{name, confpath})
			seen[confpath] = true
		}
	}

	return

}

func ampBuild(argv []string, usage string) {

	opts := optparse.Parser(
		"Usage: amp build <path> [options]\n\n    " + usage + "\n")

	args := opts.Parse(argv)

	if len(args) == 0 {
		opts.PrintUsage()
		runtime.Exit(0)
	}

	for _, info := range getNodules(args) {
		nodule.Build(info[0], info[1])
	}

}

func ampTest(argv []string, usage string) {

	opts := optparse.Parser(
		"Usage: amp test <path> [options]\n\n    " + usage + "\n")

	args := opts.Parse(argv)

	if len(args) == 0 {
		opts.PrintUsage()
		runtime.Exit(0)
	}

	for _, info := range getNodules(args) {
		nodule.Test(info[0], info[1])
	}

}
