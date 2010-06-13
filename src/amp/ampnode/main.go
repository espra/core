// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package main

import "amp/runtime"
import "fmt"

func main() {
	runtime.Init()
	fmt.Printf("Running Ampnode with %d CPUs.\n", runtime.CPUCount)
}
