// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Command genasm generates Go assembly files for packages.
package main

import (
	"fmt"
	"os"

	_ "dappui.com/cmd/genasm/blake3"
	"dappui.com/cmd/genasm/pkg"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Usage: genasm PACKAGES ...")
		os.Exit(0)
	}
	pkg.Generate(os.Args[1:])
}
