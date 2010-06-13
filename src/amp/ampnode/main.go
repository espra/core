// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package main

import "amp/runtime"
import "fmt"
import "os"

var AMPIFY_ROOT string

func main() {

	// Check if ``$AMPIFY_ROOT`` has been set.
	AMPIFY_ROOT = os.Getenv("AMPIFY_ROOT")
	if AMPIFY_ROOT == "" {
		fmt.Printf(
			"ERROR: The AMPIFY_ROOT environment variable hasn't been set.\n")
		os.Exit(1)
	}

	// Run Ampnode on multiple processors if possible.
	runtime.Init()
	fmt.Printf("Running Ampnode with %d CPUs.\n", runtime.CPUCount)

}
