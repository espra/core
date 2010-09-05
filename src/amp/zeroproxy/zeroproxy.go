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
	contentType      = "Content-Type"
	contentLength    = "Content-Length"
	redirectHTML     = `Please <a href="%s">click here if your browser doesn't redirect</a> automatically.`
	redirectURL      = "%s%s"
	redirectURLQuery = "%s%s?%s"
	textHTML         = "text/html"
)

var (
	debugMode  bool
	remoteAddr string
	remoteHost string
)

var (
	error502       = []byte(`<!DOCTYPE html>
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
	error502Length = fmt.Sprintf("%d", len(error502))
)

type Redirector struct {
	url string
}

func (redirector *Redirector) ServeHTTP(conn *http.Conn, req *http.Request) {

	var url string
	if len(req.URL.RawQuery) > 0 {
		url = fmt.Sprintf(redirectURLQuery, redirector.url, req.URL.Path, req.URL.RawQuery)
	} else {
		url = fmt.Sprintf(redirectURL, redirector.url, req.URL.Path)
	}

	if len(url) == 0 {
		url = "/"
	}

	conn.SetHeader("Location", url)
	conn.WriteHeader(http.StatusMovedPermanently)
	fmt.Fprintf(conn, redirectHTML, url)

}


type Proxy struct {
	tlsMode bool
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

func serveError502(conn *http.Conn) {
	conn.WriteHeader(http.StatusBadGateway)
	conn.SetHeader(contentType, textHTML)
	conn.SetHeader(contentLength, error502Length)
	conn.Write(error502)
}

func main() {

	opts := optparse.Parser(
		"Usage: zeroproxy <zerolive-config.yaml> [options]\n",
		"zeroproxy 0.0.0")

	host := opts.String([]string{"--host"}, "",
		"the host to bind the HTTPS server to", "HOST")

	port := opts.Int([]string{"--port"}, 9443,
		"the port to bind the HTTPS server to [default: 9443]", "PORT")

	certFile := opts.String([]string{"--cert-file"}, "",
		"the path to the TLS certificate for the HTTPS server", "FILE")

	httpHost := opts.String([]string{"--http-host"}, "",
		"the host to bind the HTTP server to", "HOST")

	httpPort := opts.Int([]string{"--http-port"}, 9080,
		"the port to bind the HTTP server to [default: 9080]", "PORT")

	httpRedirect := opts.String([]string{"--redirect"},
		"https://localhost:9443",
		"the redirect for HTTP requests [default: https://localhost:9443]",
		"URL")

	remoteHost := opts.String([]string{"--gae-host"}, "localhost",
		"the App Engine host to connect to [default: localhost]", "HOST")

	remotePort := opts.Int([]string{"--gae-port"}, 8080,
		"the App Engine port to connect to [default: 8080]", "PORT")

	tlsMode := opts.Bool([]string{"--tls-proxy"}, false,
		"proxy to App Engine using TLS [default: false]")

	debug := opts.Bool([]string{"-d", "--debug"}, false,
		"enable debug mode")

	os.Args[0] = "zeroproxy"
	args := opts.Parse(os.Args)

	var liveConfig string

	if len(args) >= 1 {
		if args[0] == "help" {
			opts.PrintUsage()
			os.Exit(0)
		}
		liveConfig = args[0]
	} else {
		opts.PrintUsage()
		os.Exit(0)
	}

	_ = liveConfig
	_ = certFile

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

	httpAddr := fmt.Sprintf("%s:%d", *httpHost, *httpPort)
	httpListener, err := net.Listen("tcp", httpAddr)
	if err != nil {
		fmt.Printf("Cannot listen on %s: %v\n", httpAddr, err)
		os.Exit(1)
	}

	var addrURL, httpAddrURL string

	if len(*host) == 0 {
		addrURL = fmt.Sprintf("https://localhost:%d", *port)
	} else {
		addrURL = fmt.Sprintf("https://%s:%d", *host, *port)
	}

	if len(*httpHost) == 0 {
		httpAddrURL = fmt.Sprintf("http://localhost:%d", *httpPort)
	} else {
		httpAddrURL = fmt.Sprintf("http://%s:%d", *httpHost, *httpPort)
	}

	fmt.Printf("Running zeroproxy with %d CPUs:\n", runtime.CPUCount)
	fmt.Printf("* Frontend proxy running on %s\n", addrURL)
	fmt.Printf("* HTTP Redirector running on %s -> %s\n", httpAddrURL, *httpRedirect)

	redirector := &Redirector{url: *httpRedirect}
	http.Serve(httpListener, redirector)

	proxy := &Proxy{
		tlsMode: *tlsMode,
	}
	http.Serve(listener, proxy)

}
