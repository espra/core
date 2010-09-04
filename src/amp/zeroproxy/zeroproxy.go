// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// Zero Proxy
// ==========
//
// The ``zeroproxy`` app proxies requests to:
//
// 1. Google App Engine -- this is needed as App Engine doesn't yet support
//    HTTPS requests on custom domains.
//
// 2. The ``zerolive`` app -- which, in turn, interacts with Redis and Keyspace
//    nodes.
//
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

const (
	ContentType   = "Content-Type"
	ContentLength = "Content-Length"
	TextHTML      = "text/html"
)

var (
	Error502       = []byte(`<!DOCTYPE html>
<html>
<head>
<title>DoctError!</title>
<link href="//fonts.googleapis.com/css?family=Josefin+Sans+Std+Light:regular" rel="stylesheet" type="text/css" >
<style>
body {
font-family: 'Josefin Sans Std Light', serif;
font-size: 28px;
font-weight: 400;
line-height: 40px;
background: #ebf3f6;
padding: 50px;
margin: 0;
}
</style>
</head>
<body>
Our servers ran into DoctError and are too frightened to continue.
<br/>
Help fight evil by waiting a moment and trying again. Thanks!
</body></html>`)
	Error502Length = fmt.Sprintf("%d", len(Error502))
)

var (
	debugMode  bool
	remoteAddr string
	remoteHost string
)

type Proxy struct{}

func serveError502(conn *http.Conn) {
	conn.WriteHeader(502)
	conn.SetHeader(ContentType, TextHTML)
	conn.SetHeader(ContentLength, Error502Length)
	conn.Write(Error502)
}

func (proxy *Proxy) ServeHTTP(conn *http.Conn, req *http.Request) {

	// Open a connection to the App Engine server.
	aeconn, err := net.Dial("tcp", "", remoteAddr)
	if err != nil {
		if debugMode {
			fmt.Printf("Couldn't connect to remote %s: %v\n", remoteHost, err)
		}
		serveError502(conn)
		return
	}

	ae := tls.Client(aeconn, tlsconf.Config)
	defer ae.Close()

	// Modify the request Host: header.
	req.Host = remoteHost

	// Send the request to the App Engine server.
	err = req.Write(ae)
	if err != nil {
		if debugMode {
			fmt.Printf("Error writing to App Engine: %v\n", err)
		}
		serveError502(conn)
		return
	}

	// Parse the response from App Engine.
	resp, err := http.ReadResponse(bufio.NewReader(ae), req.Method)
	if err != nil {
		if debugMode {
			fmt.Printf("Error parsing response from App Engine: %v\n", err)
		}
		serveError502(conn)
		return
	}

	// Read the full response body.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if debugMode {
			fmt.Printf("Error reading response from App Engine: %v\n", err)
		}
		serveError502(conn)
		resp.Body.Close()
		return
	}

	// Set the received headers back to the initial connection.
	for k, v := range resp.Header {
		conn.SetHeader(k, v)
	}

	// Write the response body back to the initial connection.
	resp.Body.Close()
	conn.WriteHeader(resp.StatusCode)
	conn.Write(body)

}

func main() {

	opts := optparse.Parser("Usage: zeroproxy [options]\n", "zeroproxy 0.0.0")

	host := opts.String([]string{"--host"}, "localhost",
		"the host to bind to [default: localhost]")

	port := opts.Int([]string{"--port"}, 8010,
		"the port to bind to [default: 8010]")

	remoteHost := opts.String([]string{"--remote-host"}, "localhost",
		"the remote host to connect to [default: localhost]")

	remotePort := opts.Int([]string{"--remote-port"}, 8080,
		"the remote port to connect to [default: 8080]")

	tlsMode := opts.Bool([]string{"--tls"}, false,
		"enable TLS (SSL) mode when connecting to the remote host")

	debug := opts.Bool([]string{"--debug"}, false,
		"enable debug mode")

	os.Args[0] = "zeroproxy"
	args := opts.Parse(os.Args)

	if len(args) >= 1 {
		if args[0] == "help" {
			opts.PrintUsage()
			os.Exit(1)
		}
	}

	_ = tlsMode

	// Initialise the Ampify runtime -- which will run ``zeroproxy`` on multiple
	// processors if possible.
	runtime.Init()

	// Initialise the TLS config.
	tlsconf.Init()

	debugMode = *debug
	remoteAddr = fmt.Sprintf("%s:%d", *remoteHost, *remotePort)
	addr := fmt.Sprintf("%s:%d", *host, *port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Cannot listen on %s: %v\n", addr, err)
		os.Exit(1)
	}

	fmt.Printf("Running zeroproxy with %d CPUs on %s\n",
		runtime.CPUCount, addr)

	proxy := &Proxy{}
	http.Serve(listener, proxy)

}
