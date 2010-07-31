// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// The ``hubproxy`` command proxies requests to the Amp Hub backend. This is
// needed as the backend is currently running on top of Google App Engine and
// it doesn't support HTTPS requests on custom domains yet.
package main

import "amp/runtime"
import "amp/optparse"
import "fmt"
import "os"

func main() {

	opts := optparse.Parser("Usage: hubproxy [options]\n", "hubproxy 0.0.0")
	port := opts.Int([]string{"-p", "--port"}, 8010, "specify the port number to use [default: 8010]")
	host := opts.String([]string{"--host"}, "localhost", "specify the host to bind to")

	os.Args[0] = "hubproxy"
	args := opts.Parse(os.Args)

	if len(args) >= 1 {
		if args[0] == "help" {
			opts.PrintUsage()
			os.Exit(1)
		}
	}

	// Run the hubproxy on multiple processors if possible.
	runtime.Init()
	fmt.Printf("Running hubproxy with %d CPUs on %s:%d\n", runtime.CPUCount, *host, *port)

}
