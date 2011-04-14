// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// Tent Proxy
// ==========
//
// The ``tentproxy`` app proxies requests to:
//
// 1. A ``tentapp`` instance running on Google App Engine -- this is needed as
//    App Engine doesn't yet support HTTPS requests on custom domains.
//
// 2. The ``airbeam`` app -- which, in turn, interacts with Redis and Keyspace
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
)

var (
	error502       = []byte(`<!DOCTYPE html>
<html>
<head>
<title>DoctError! Couldn't connect to an upstream server.</title>
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

func (redirector *Redirector) ServeHTTP(conn http.ResponseWriter, req *http.Request) {

	var url string
	if len(req.URL.RawQuery) > 0 {
		url = fmt.Sprintf(redirectURLQuery, redirector.url, req.URL.Path, req.URL.RawQuery)
	} else {
		url = fmt.Sprintf(redirectURL, redirector.url, req.URL.Path)
	}

	if len(url) == 0 {
		url = "/"
	}

	conn.Header().Set("Location", url)
	conn.WriteHeader(http.StatusMovedPermanently)
	fmt.Fprintf(conn, redirectHTML, url)
	logRequest(http.StatusMovedPermanently, req.Host, conn, req)

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

func (frontend *Frontend) ServeHTTP(conn http.ResponseWriter, req *http.Request) {

	if frontend.enforceHost && req.Host != frontend.officialHost {
		conn.Header().Set("Location", frontend.officialRedirectURL)
		conn.WriteHeader(http.StatusMovedPermanently)
		conn.Write(frontend.officialRedirectHTML)
		logRequest(http.StatusMovedPermanently, req.Host, conn, req)
		return
	}

	originalHost := req.Host

	// Open a connection to the App Engine server.
	gaeConn, err := net.Dial("tcp", frontend.gaeAddr)
	if err != nil {
		if debugMode {
			fmt.Printf("Couldn't connect to remote %s: %v\n", frontend.gaeHost, err)
		}
		serveError502(conn, originalHost, req)
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
		serveError502(conn, originalHost, req)
		return
	}

	// Parse the response from App Engine.
	resp, err := http.ReadResponse(bufio.NewReader(gae), req.Method)
	if err != nil {
		if debugMode {
			fmt.Printf("Error parsing response from App Engine: %v\n", err)
		}
		serveError502(conn, originalHost, req)
		return
	}

	// Read the full response body.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if debugMode {
			fmt.Printf("Error reading response from App Engine: %v\n", err)
		}
		serveError502(conn, originalHost, req)
		resp.Body.Close()
		return
	}

	// Get the header.
	headers := conn.Header()

	// Set the received headers back to the initial connection.
	for k, values := range resp.Header {
		for _, v := range values {
			headers.Add(k, v)
		}
	}

	// Write the response body back to the initial connection.
	resp.Body.Close()
	conn.WriteHeader(resp.StatusCode)
	conn.Write(body)

	logRequest(resp.StatusCode, originalHost, conn, req)

}

func logRequest(status int, host string, conn http.ResponseWriter, request *http.Request) {
	var ip string
	splitPoint := strings.LastIndex(request.RemoteAddr, ":")
	if splitPoint == -1 {
		ip = request.RemoteAddr
	} else {
		ip = request.RemoteAddr[0:splitPoint]
	}
	logging.Info("fe", status, request.Method, host, request.RawURL,
		ip, request.UserAgent, request.Referer)
}

func filterRequestLog(record *logging.Record) (write bool, data []interface{}) {
	itemLength := len(record.Items)
	if itemLength > 1 {
		switch record.Items[0].(type) {
		case string:
			if record.Items[0].(string) == "fe" {
				return true, record.Items[1 : itemLength-2]
			}
		}
	}
	return true, data
}

func serveError502(conn http.ResponseWriter, host string, request *http.Request) {
	headers := conn.Header()
	headers.Set(contentType, textHTML)
	headers.Set(contentLength, error502Length)
	conn.WriteHeader(http.StatusBadGateway)
	conn.Write(error502)
	logRequest(http.StatusBadGateway, host, conn, request)
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

	// masterHosts := opts.BoolConfig("master-hosts", false,
	// 	"disable the HTTP Redirector [default: false]")

	keyspaceMaster := opts.BoolConfig("master", false,
		"disable the HTTP Redirector [default: false]")

	_ = keyspaceMaster

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
			runtime.Exit(0)
		}
		instanceDirectory = path.Clean(args[0])
	} else {
		opts.PrintUsage()
		runtime.Exit(0)
	}

	rootInfo, err := os.Stat(instanceDirectory)
	if err == nil {
		if !rootInfo.IsDirectory() {
			runtime.Error("ERROR: %q is not a directory\n", instanceDirectory)
		}
	} else {
		runtime.Error("ERROR: %s\n", err)
	}

	configPath := path.Join(instanceDirectory, "ampzero.yaml")
	_, err = os.Stat(configPath)
	if err == nil {
		err = opts.ParseConfig(configPath, os.Args)
		if err != nil {
			runtime.Error("ERROR: %s\n", err)
		}
	}

	logPath := path.Join(instanceDirectory, "log")
	err = os.MkdirAll(logPath, 0755)
	if err != nil {
		runtime.Error("ERROR: %s\n", err)
	}

	runPath := path.Join(instanceDirectory, "run")
	err = os.MkdirAll(runPath, 0755)
	if err != nil {
		runtime.Error("ERROR: %s\n", err)
	}

	_, err = runtime.GetLock(runPath, "ampzero")
	if err != nil {
		runtime.Error("ERROR: Couldn't successfully acquire a process lock:\n\n\t%s\n\n", err)
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
			runtime.Exit(1)
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
		runtime.Error("ERROR: Cannot listen on %s: %v\n", frontendAddr, err)
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
			runtime.Error("ERROR: Couldn't load certificate/key pair: %s\n", err)
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
			runtime.Error("ERROR: Cannot listen on %s: %v\n", httpAddr, err)
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
		runtime.Error("ERROR: Unknown log rotation format %q\n", *logRotate)
	}

	if !*noConsoleLog {
		logging.AddConsoleLogger()
		logging.AddFilter(filterRequestLog)
	}

	_, err = logging.AddFileLogger("ampzero", logPath, rotate)
	if err != nil {
		runtime.Error("ERROR: Couldn't initialise logfile: %s\n", err)
	}

	fmt.Printf("Running ampzero with %d CPUs:\n", runtime.CPUCount)

	if !*noRedirect {
		redirector := &Redirector{url: *redirectURL}
		go func() {
			err = http.Serve(httpListener, redirector)
			if err != nil {
				runtime.Error("ERROR serving HTTP Redirector: %s\n", err)
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
		runtime.Error("ERROR serving Frontend Server: %s\n", err)
	}

}
