// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package main

import (
	"amp/optparse"
)

func main() {

	commands := map[string]func([]string, string){
		"build":    ampBuild,
		"frontend": ampFrontend,
		"init":     ampInit,
		"master":   ampMaster,
		"node":     ampNode,
		"pull":     ampPull,
		"push":     ampPush,
		"review":   ampReview,
		"run":      ampRun,
		"test":     ampTest,
	}

	usage := map[string]string{
		"build":    "build an amp nodule",
		"frontend": "run an amp frontend",
		"init":     "initialise an instance",
		"master":   "run an amp master node",
		"node":     "run an amp node",
		"pull":     "pull from an amp master node",
		"push":     "push to an amp master node",
		"review":   "review an amp nodule",
		"run":      "run a combined single-server instance",
		"test":     "test an amp nodule",
		"version":  "show the version number and exit",
	}

	optparse.Subcommands("amp", "amp 0.0.0", commands, usage)

}
