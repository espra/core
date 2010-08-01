// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package optparse

import (
	"testing"
)

func TestVersion(t *testing.T) {

	opts := Parser("Usage: test", "version string")
	if opts.Version != "version string" {
		t.Error("Version string wasn't set.\n")
	}

}

func TestVersionNotSet(t *testing.T) {

	opts := Parser("Usage: test")
	if opts.Version != "" {
		t.Error("Version string was unexpectedly set.\n")
	}

}

func TestFlags(t *testing.T) {

	opts := Parser("Usage: test", "version string")
	port := opts.Int([]string{"-p", "--port"}, 8010, "specify the port number to use")
	host := opts.String([]string{"--host"}, "localhost", "specify the host to bind to")

	args := opts.Parse([]string{"testapp", "-p", "8040", "--host", "asktav.com"})

	if len(args) >= 1 {
		t.Error("Got unexpected arguments back.\n")
	}

	if *port != 8040 {
		t.Error("Got an invalid value for the --port parameter.\n")
	}

	if *host != "asktav.com" {
		t.Error("Got an invalid value for the --host parameter.\n")
	}

}

func TestArgs(t *testing.T) {

	opts := Parser("Usage: test", "version string")
	opts.Int([]string{"-p", "--port"}, 8010, "specify the port number to use")
	opts.String([]string{"--host"}, "localhost", "specify the host to bind to")

	args := opts.Parse([]string{"testapp", "foo1", "foo2"})

	if len(args) != 2 {
		t.Error("Got an invalid number of arguments.\n")
		return
	}

	if args[0] != "foo1" {
		t.Error("Got an invalid first argument.\n")
	}

	if args[1] != "foo2" {
		t.Error("Got an invalid second argument.\n")
	}

}
