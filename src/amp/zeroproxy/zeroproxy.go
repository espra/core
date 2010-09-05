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
	"amp/optparse"
	"amp/runtime"
	"amp/tlsconf"
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"http"
	"io/ioutil"
	"net"
	"os"
	"time"
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
	debugMode bool
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
	gaeAddr              string
	gaeHost              string
	gaeTLS               bool
	officialHost         string
	officialRedirectURL  string
	officialRedirectHTML []byte
	enforceHost          bool
}

func (proxy *Proxy) ServeHTTP(conn *http.Conn, req *http.Request) {

	if proxy.enforceHost && req.Host != proxy.officialHost {
		conn.SetHeader("Location", proxy.officialRedirectURL)
		conn.WriteHeader(http.StatusMovedPermanently)
		conn.Write(proxy.officialRedirectHTML)
		return
	}

	// Open a connection to the App Engine server.
	gaeConn, err := net.Dial("tcp", "", proxy.gaeAddr)
	if err != nil {
		if debugMode {
			fmt.Printf("Couldn't connect to remote %s: %v\n", proxy.gaeHost, err)
		}
		serveError502(conn)
		return
	}

	var gae net.Conn

	if proxy.gaeTLS {
		gae = tls.Client(gaeConn, tlsconf.Config)
		defer gae.Close()
	} else {
		gae = gaeConn
	}

	// Modify the request Host: header.
	req.Host = proxy.gaeHost

	// Send the request to the App Engine server.
	err = req.Write(gae)
	if err != nil {
		if debugMode {
			fmt.Printf("Error writing to App Engine: %v\n", err)
		}
		serveError502(conn)
		return
	}

	// Parse the response from App Engine.
	resp, err := http.ReadResponse(bufio.NewReader(gae), req.Method)
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

	proxyHost := opts.String([]string{"--host"}, "",
		"the host to bind the Proxy Frontend to", "HOST")

	proxyPort := opts.Int([]string{"--port"}, 9443,
		"the port to bind the Proxy Frontend to [default: 9443]", "PORT")

	certFile := opts.String([]string{"--cert-file"}, "",
		"the path to the TLS certificate for the Proxy Frontend", "PATH")

	keyFile := opts.String([]string{"--key-file"}, "",
		"the path to the TLS key for the Proxy Frontend", "PATH")

	officialHost := opts.String([]string{"--official"}, "",
		"if specified, limit the Proxy Frontend to just this host", "HOST")

	httpHost := opts.String([]string{"--http-host"}, "",
		"the host to bind the HTTP Redirector to", "HOST")

	httpPort := opts.Int([]string{"--http-port"}, 9080,
		"the port to bind the HTTP Redirector to [default: 9080]", "PORT")

	httpRedirect := opts.String([]string{"--redirect"},
		"https://localhost:9443",
		"the redirect for HTTP requests [default: https://localhost:9443]",
		"URL")

	disableHTTP := opts.Bool([]string{"--disable-http"}, false,
		"disable the HTTP redirector")

	gaeHost := opts.String([]string{"--gae-host"}, "localhost",
		"the App Engine host to connect to [default: localhost]", "HOST")

	gaePort := opts.Int([]string{"--gae-port"}, 8080,
		"the App Engine port to connect to [default: 8080]", "PORT")

	gaeTLS := opts.Bool([]string{"--gae-tls"}, false,
		"use TLS when connecting to App Engine [default: false]")

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

	var exitProcess bool

	if len(*certFile) == 0 {
		fmt.Printf("ERROR: The required --cert-file parameter hasn't been provided.\n")
		exitProcess = true
	}

	if len(*keyFile) == 0 {
		fmt.Printf("ERROR: The required --key-file parameter hasn't been provided.\n")
		exitProcess = true
	}

	if exitProcess {
		os.Exit(1)
	}

	_ = liveConfig

	// Initialise the Ampify runtime -- which will run ``zeroproxy`` on multiple
	// processors if possible.
	runtime.Init()

	// Initialise the TLS config.
	tlsconf.Init()

	debugMode = *debug
	gaeAddr := fmt.Sprintf("%s:%d", *gaeHost, *gaePort)

	proxyAddr := fmt.Sprintf("%s:%d", *proxyHost, *proxyPort)
	proxyConn, err := net.Listen("tcp", proxyAddr)
	if err != nil {
		fmt.Printf("Cannot listen on %s: %v\n", proxyAddr, err)
		os.Exit(1)
	}

	tlsConfig := &tls.Config{
		NextProtos: []string{"http/1.1"},
		Rand:       rand.Reader,
		Time:       time.Seconds,
	}

	tlsConfig.Certificates = make([]tls.Certificate, 1)
	tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		fmt.Printf("Error loading certificate/key pair: %s\n", err)
		os.Exit(1)
	}

	proxyListener := tls.NewListener(proxyConn, tlsConfig)

	var httpAddr string
	var httpListener net.Listener

	if !*disableHTTP {
		httpAddr = fmt.Sprintf("%s:%d", *httpHost, *httpPort)
		httpListener, err = net.Listen("tcp", httpAddr)
		if err != nil {
			fmt.Printf("Cannot listen on %s: %v\n", httpAddr, err)
			os.Exit(1)
		}
	}

	var enforceHost bool
	var officialRedirectURL string
	var officialRedirectHTML []byte

	if len(*officialHost) != 0 {
		enforceHost = true
		officialRedirectURL = "https://" + *officialHost + "/"
		officialRedirectHTML = []byte(fmt.Sprintf(redirectHTML, officialRedirectURL))
	}

	var proxyAddrURL, httpAddrURL string

	if len(*proxyHost) == 0 {
		proxyAddrURL = fmt.Sprintf("https://localhost:%d", *proxyPort)
	} else {
		proxyAddrURL = fmt.Sprintf("https://%s:%d", *proxyHost, *proxyPort)
	}

	if len(*httpHost) == 0 {
		httpAddrURL = fmt.Sprintf("http://localhost:%d", *httpPort)
	} else {
		httpAddrURL = fmt.Sprintf("http://%s:%d", *httpHost, *httpPort)
	}

	fmt.Printf("Running zeroproxy with %d CPUs:\n", runtime.CPUCount)

	if !*disableHTTP {
		redirector := &Redirector{url: *httpRedirect}
		go func() {
			err = http.Serve(httpListener, redirector)
			if err != nil {
				fmt.Printf("ERROR serving HTTP Redirector: %s\n", err)
				os.Exit(1)
			}
		}()
		fmt.Printf("* HTTP Redirector running on %s -> %s\n", httpAddrURL, *httpRedirect)
	}

	proxy := &Proxy{
		gaeAddr:              gaeAddr,
		gaeHost:              *gaeHost,
		gaeTLS:               *gaeTLS,
		officialHost:         *officialHost,
		officialRedirectURL:  officialRedirectURL,
		officialRedirectHTML: officialRedirectHTML,
		enforceHost:          enforceHost,
	}

	fmt.Printf("* Frontend Proxy running on %s\n", proxyAddrURL)

	err = http.Serve(proxyListener, proxy)
	if err != nil {
		fmt.Printf("ERROR serving Frontend Proxy: %s\n", err)
		os.Exit(1)
	}

}
