// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// Ampify Zero
// ===========
//
// The ``ampzero`` app proxies requests to:
//
// 1. Google App Engine -- this is needed as App Engine doesn't yet support
//    HTTPS requests on custom domains.
//
// 2. The ``zerolive`` app -- which, in turn, interacts with Redis and Keyspace
//    nodes.
//
package main

import (
	"amp/logging"
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
	"path"
	"strings"
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
	log       *logging.Logger
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
	logRequest(http.StatusMovedPermanently, conn)

}

type Frontend struct {
	gaeAddr              string
	gaeHost              string
	gaeTLS               bool
	officialHost         string
	officialRedirectURL  string
	officialRedirectHTML []byte
	enforceHost          bool
}

func (frontend *Frontend) ServeHTTP(conn *http.Conn, req *http.Request) {

	if frontend.enforceHost && req.Host != frontend.officialHost {
		conn.SetHeader("Location", frontend.officialRedirectURL)
		conn.WriteHeader(http.StatusMovedPermanently)
		conn.Write(frontend.officialRedirectHTML)
		logRequest(http.StatusMovedPermanently, conn)
		return
	}

	// Open a connection to the App Engine server.
	gaeConn, err := net.Dial("tcp", "", frontend.gaeAddr)
	if err != nil {
		if debugMode {
			fmt.Printf("Couldn't connect to remote %s: %v\n", frontend.gaeHost, err)
		}
		serveError502(conn)
		return
	}

	var gae net.Conn

	if frontend.gaeTLS {
		gae = tls.Client(gaeConn, tlsconf.Config)
		defer gae.Close()
	} else {
		gae = gaeConn
	}

	// Modify the request Host: header.
	req.Host = frontend.gaeHost

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

	logRequest(resp.StatusCode, conn)

}

func logRequest(status int, conn *http.Conn) {
	var ip string
	splitPoint := strings.LastIndex(conn.RemoteAddr, ":")
	if splitPoint == -1 {
		ip = conn.RemoteAddr
	} else {
		ip = conn.RemoteAddr[0:splitPoint]
	}
	request := conn.Req
	log.Log("fe", status, request.Method, request.Host, request.RawURL,
		ip, request.UserAgent, request.Referer)
}

func serveError502(conn *http.Conn) {
	conn.WriteHeader(http.StatusBadGateway)
	conn.SetHeader(contentType, textHTML)
	conn.SetHeader(contentLength, error502Length)
	conn.Write(error502)
	logRequest(http.StatusBadGateway, conn)
}

func main() {

	opts := optparse.Parser(
		"Usage: ampzero </path/to/instance/directory> [options]\n",
		"ampzero 0.0.0")

	debug := opts.Bool([]string{"-d", "--debug"}, false,
		"enable debug mode")

	frontendHost := opts.StringConfig("frontend-host", "",
		"the host to bind the Frontend Server to")

	frontendPort := opts.IntConfig("frontend-port", 9040,
		"the port to bind the Frontend Server to [default: 9040]")

	frontendTLS := opts.BoolConfig("frontend-tls", false,
		"use TLS (HTTPS) for the Frontend Server [default: false]")

	certFile := opts.StringConfig("cert-file", "cert/frontend.cert",
		"the path to the TLS certificate [default: cert/frontend.cert]")

	keyFile := opts.StringConfig("key-file", "cert/frontend.key",
		"the path to the TLS key [default: cert/frontend.key]")

	officialHost := opts.StringConfig("official-host", "",
		"if set, limit the Frontend Server to the specified host")

	noRedirect := opts.BoolConfig("no-redirect", false,
		"disable the HTTP Redirector [default: false]")

	httpHost := opts.StringConfig("http-host", "",
		"the host to bind the HTTP Redirector to")

	httpPort := opts.IntConfig("http-port", 9080,
		"the port to bind the HTTP Redirector to [default: 9080]")

	redirectURL := opts.StringConfig("redirect-url", "",
		"the URL that the HTTP Redirector redirects to")

	gaeHost := opts.StringConfig("gae-host", "localhost",
		"the App Engine host to connect to [default: localhost]")

	gaePort := opts.IntConfig("gae-port", 8080,
		"the App Engine port to connect to [default: 8080]")

	gaeTLS := opts.BoolConfig("gae-tls", false,
		"use TLS when connecting to App Engine [default: false]")

	logRotate := opts.StringConfig("log-rotate", "never",
		"specify one of 'hourly', 'daily' or 'never' [default: never]")

	noConsoleLog := opts.BoolConfig("no-console-log", false,
		"disable logging to stdout/stderr [default: false]")

	os.Args[0] = "ampzero"
	args := opts.Parse(os.Args)

	var instanceDirectory string

	if len(args) >= 1 {
		if args[0] == "help" {
			opts.PrintUsage()
			os.Exit(0)
		}
		instanceDirectory = path.Clean(args[0])
	} else {
		opts.PrintUsage()
		os.Exit(0)
	}

	rootInfo, err := os.Stat(instanceDirectory)
	if err == nil {
		if !rootInfo.IsDirectory() {
			fmt.Printf("ERROR: %q is not a directory\n", instanceDirectory)
			os.Exit(1)
		}
	} else {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	configPath := path.Join(instanceDirectory, "ampzero.yaml")
	_, err = os.Stat(configPath)
	if err == nil {
		err = opts.ParseConfig(configPath, os.Args)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
	}

	logPath := path.Join(instanceDirectory, "log")
	err = os.MkdirAll(logPath, 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	runPath := path.Join(instanceDirectory, "run")
	err = os.MkdirAll(runPath, 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	if runtime.AcquireLock(path.Join(runPath, "ampzero.lock")) != 0 {
		fmt.Printf("ERROR: Another ampzero process is already running in %s\n", instanceDirectory)
		os.Exit(1)
	}

	go runtime.CreatePidFile(path.Join(runPath, "ampzero.pid"))

	if *frontendTLS {
		var exitProcess bool
		if len(*certFile) == 0 {
			fmt.Printf("ERROR: The cert-file config value hasn't been specified.\n")
			exitProcess = true
		}
		if len(*keyFile) == 0 {
			fmt.Printf("ERROR: The key-file config value hasn't been specified.\n")
			exitProcess = true
		}
		if exitProcess {
			os.Exit(1)
		}
	}

	// Initialise the Ampify runtime -- which will run ``ampzero`` on multiple
	// processors if possible.
	runtime.Init()

	// Initialise the TLS config.
	tlsconf.Init()

	debugMode = *debug
	gaeAddr := fmt.Sprintf("%s:%d", *gaeHost, *gaePort)

	frontendAddr := fmt.Sprintf("%s:%d", *frontendHost, *frontendPort)
	frontendConn, err := net.Listen("tcp", frontendAddr)
	if err != nil {
		fmt.Printf("ERROR: Cannot listen on %s: %v\n", frontendAddr, err)
		os.Exit(1)
	}

	var frontendListener net.Listener

	if *frontendTLS {
		certPath := path.Join(instanceDirectory, *certFile)
		keyPath := path.Join(instanceDirectory, *keyFile)
		tlsConfig := &tls.Config{
			NextProtos: []string{"http/1.1"},
			Rand:       rand.Reader,
			Time:       time.Seconds,
		}
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			fmt.Printf("ERROR: Couldn't load certificate/key pair: %s\n", err)
			os.Exit(1)
		}
		frontendListener = tls.NewListener(frontendConn, tlsConfig)
	} else {
		frontendListener = frontendConn
	}

	var enforceHost bool
	var officialRedirectURL string
	var officialRedirectHTML []byte

	if len(*officialHost) != 0 {
		enforceHost = true
		if *frontendTLS {
			officialRedirectURL = "https://" + *officialHost + "/"
		} else {
			officialRedirectURL = "http://" + *officialHost + "/"
		}
		officialRedirectHTML = []byte(fmt.Sprintf(redirectHTML, officialRedirectURL))
	}

	var frontendScheme, frontendAddrURL, httpAddrURL string

	if *frontendTLS {
		frontendScheme = "https://"
	} else {
		frontendScheme = "http://"
	}

	if len(*frontendHost) == 0 {
		frontendAddrURL = fmt.Sprintf("%slocalhost:%d", frontendScheme, *frontendPort)
	} else {
		frontendAddrURL = fmt.Sprintf("%s%s:%d", frontendScheme, *frontendHost, *frontendPort)
	}

	if len(*httpHost) == 0 {
		httpAddrURL = fmt.Sprintf("http://localhost:%d", *httpPort)
	} else {
		httpAddrURL = fmt.Sprintf("http://%s:%d", *httpHost, *httpPort)
	}

	var httpAddr string
	var httpListener net.Listener

	if !*noRedirect {
		if *redirectURL == "" {
			*redirectURL = frontendAddrURL
		}
		httpAddr = fmt.Sprintf("%s:%d", *httpHost, *httpPort)
		httpListener, err = net.Listen("tcp", httpAddr)
		if err != nil {
			fmt.Printf("ERROR: Cannot listen on %s: %v\n", httpAddr, err)
			os.Exit(1)
		}
	}

	var rotate int

	switch *logRotate {
	case "daily":
		rotate = logging.RotateDaily
	case "hourly":
		rotate = logging.RotateHourly
	case "never":
		rotate = logging.RotateNever
	default:
		fmt.Printf("ERROR: Unknown log rotation format %q\n", *logRotate)
		os.Exit(1)
	}

	log, err = logging.New("ampzero", logPath, rotate, !*noConsoleLog)
	if err != nil {
		fmt.Printf("ERROR: Couldn't initialise logfile: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Running ampzero with %d CPUs:\n", runtime.CPUCount)

	if !*noRedirect {
		redirector := &Redirector{url: *redirectURL}
		go func() {
			err = http.Serve(httpListener, redirector)
			if err != nil {
				fmt.Printf("ERROR serving HTTP Redirector: %s\n", err)
				os.Exit(1)
			}
		}()
		fmt.Printf("* HTTP Redirector running on %s -> %s\n", httpAddrURL, *redirectURL)
	}

	frontend := &Frontend{
		gaeAddr:              gaeAddr,
		gaeHost:              *gaeHost,
		gaeTLS:               *gaeTLS,
		officialHost:         *officialHost,
		officialRedirectURL:  officialRedirectURL,
		officialRedirectHTML: officialRedirectHTML,
		enforceHost:          enforceHost,
	}

	fmt.Printf("* Frontend Server running on %s\n", frontendAddrURL)

	err = http.Serve(frontendListener, frontend)
	if err != nil {
		fmt.Printf("ERROR serving Frontend Server: %s\n", err)
		os.Exit(1)
	}

}
