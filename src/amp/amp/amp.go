// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package main

import (
	"amp/optparse"
	"amp/runtime"
	"fmt"
	"os"
)

func main() {

	opts := optparse.Parser(
		"Usage: amp <command> [options]\n",
		"amp 0.0.0")

	debug := opts.Bool([]string{"-d", "--debug"}, false,
		"enable debug mode")

	os.Args[0] = "amp"
	args := opts.Parse(os.Args)

	// Initialise the Ampify runtime -- which will run ``amp`` on multiple
	// processors if possible.
	runtime.Init()

	if *debug {
		fmt.Println("DEBUG: enabled")
	}

	if len(args) > 0 {
		fmt.Printf("COMMAND: %v\n", args)
	}
	
}
