// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package main

import (
	"amp/optparse"
)

func main() {

	commands := map[string]func([]string, string){
		"build":    build,
		"frontend": frontend,
		"init":     initialise,
		"node":     node,
		"pull":     pull,
		"push":     push,
		"repo":     repo,
		"review":   review,
		"run":      run,
		"store":    store,
		"test":     test,
	}

	usage := map[string]string{
		"build":    "build an amp nodule",
		"frontend": "run an amp frontend",
		"init":     "initialise an instance",
		"node":     "run an amp node",
		"pull":     "pull from an amp repo",
		"push":     "push to an amp repo",
		"repo":     "run an amp repo",
		"review":   "review an amp repo",
		"run":      "run a combined single-server instance",
		"store":    "run an amp store",
		"test":     "test an amp nodule",
		"version":  "show the version number and exit",
	}

	optparse.Subcommands("amp", "amp 0.0.0", commands, usage)

}
