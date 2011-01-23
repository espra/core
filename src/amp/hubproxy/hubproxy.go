// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// The ``hubproxy`` command proxies requests to the Amp Hub backend. This is
// needed as the backend is currently running on top of Google App Engine and
// it doesn't support HTTPS requests on custom domains yet.
package main

import (
	"amp/runtime"
	"amp/optparse"
	"amp/tlsconf"
	"bufio"
	"crypto/tls"
	"fmt"
	"http"
	"io/ioutil"
	"net"
	"os"
)

var (
	debugMode  bool
	remoteAddr string
	remoteHost string
)

type Proxy struct{}

func (proxy *Proxy) ServeHTTP(conn http.ResponseWriter, req *http.Request) {

	// Open a connection to the Hub.
	hubconn, err := net.Dial("tcp", "", remoteAddr)
	if err != nil {
		if debugMode {
			fmt.Printf("Couldn't connect to remote %s: %v\n", remoteHost, err)
		}
		return
	}

	hub := tls.Client(hubconn, tlsconf.Config)
	defer hub.Close()

	// Modify the request Host: header.
	req.Host = remoteHost

	// Send the request to the Hub.
	err = req.Write(hub)
	if err != nil {
		if debugMode {
			fmt.Printf("Error writing to the hub: %v\n", err)
		}
		return
	}

	// Parse the response from the Hub.
	resp, err := http.ReadResponse(bufio.NewReader(hub), req.Method)
	if err != nil {
		if debugMode {
			fmt.Printf("Error parsing response from the hub: %v\n", err)
		}
		return
	}

	// Set the received headers back to the initial connection.
	for k, v := range resp.Header {
		conn.SetHeader(k, v)
	}

	// Read the full response body.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if debugMode {
			fmt.Printf("Error reading response from the hub: %v\n", err)
		}
		resp.Body.Close()
		return
	}

	// Write the response body back to the initial connection.
	resp.Body.Close()
	conn.WriteHeader(resp.StatusCode)
	conn.Write(body)

}

func main() {

	opts := optparse.Parser("Usage: hubproxy [options]\n", "hubproxy 0.0.0")

	port := opts.Int([]string{"-p", "--port"}, 8010,
		"the port number to use [default: 8010]")

	host := opts.String([]string{"--host"}, "localhost",
		"the host to bind to")

	remote := opts.String([]string{"-r", "--remote"}, "ampcentral.appspot.com",
		"the remote host to connect to [default: ampcentral.appspot.com]")

	debug := opts.Bool([]string{"--debug"}, false,
		"enable debug mode")

	os.Args[0] = "hubproxy"
	args := opts.Parse(os.Args)

	if len(args) >= 1 {
		if args[0] == "help" {
			opts.PrintUsage()
			runtime.Exit(1)
		}
	}

	// Initialise the Ampify runtime -- which will run hubproxy on multiple
	// processors if possible.
	runtime.Init()

	// Initialise the TLS config.
	tlsconf.Init()

	debugMode = *debug
	remoteHost = *remote
	remoteAddr = *remote + ":443"
	addr := fmt.Sprintf("%s:%d", *host, *port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		runtime.Error("Cannot listen on %s: %v\n", addr, err)
	}

	fmt.Printf("Running hubproxy with %d CPUs on %s\n",
		runtime.CPUCount, addr)

	proxy := &Proxy{}
	http.Serve(listener, proxy)

}
