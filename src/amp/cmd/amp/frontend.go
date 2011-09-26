// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// Amp Frontend
// ============
//
package main

import (
	"amp/argo"
	"amp/livequery"
	"amp/logging"
	"amp/optparse"
	"amp/runtime"
	"amp/server"
	"amp/tlsconf"
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"http"
	"io/ioutil"
	"mime"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	acceptorKey []byte
	cookieKey   []byte
	debugMode   bool
)

// -----------------------------------------------------------------------------
// X-Live Handler
// -----------------------------------------------------------------------------

var liveChannel = make(chan []byte, 100)
var livequeryTimeout int64

func handleLiveMessages() {
	for message := range liveChannel {
		cmd := message[0]
		switch cmd {
		case 0:
			go publish(message[1:])
		case 1:
			go subscribe(message[1:])
		default:
			logging.Error("Got unexpected X-Live payload: %s", message)
		}
	}
}

// -----------------------------------------------------------------------------
// PubSub Payload Handlers
// -----------------------------------------------------------------------------

var pubsub = livequery.New()

// The publish message is of the format::
//
//     <item-id> [<key>, ...]
//
func publish(message []byte) {
	buffer := bytes.NewBuffer(message)
	decoder := &argo.Decoder{buffer}
	itemID, err := decoder.ReadString()
	if err != nil {
		logging.Error("Error decoding X-Live Publish ItemID %s %s", message, err)
		return
	}
	if itemID == "" {
		return
	}
	keys, err := decoder.ReadStringArray()
	if err != nil {
		logging.Error("Error decoding X-Live Publish Keys %s %s", message, err)
		return
	}
	pubsub.Publish(itemID, keys)
}

// The subscribe message is of the format::
//
//     <sqid> [<key>, ...] [<key>, ...]
//
// The two sets of keys are assumed to be disjoint sets, i.e. they have no
// elements in common.
func subscribe(message []byte) {
	buffer := bytes.NewBuffer(message)
	decoder := &argo.Decoder{buffer}
	sqid, err := decoder.ReadString()
	if err != nil {
		logging.Error("Error decoding X-Live Subscribe SQID %s %s", message, err)
		return
	}
	if sqid == "" {
		return
	}
	keys1, err := decoder.ReadStringArray()
	if err != nil {
		logging.Error("Error decoding X-Live Subscribe Keys-1 %s %s", message, err)
		return
	}
	var keys2 []string
	if buffer.Len() > 0 {
		keys2, err = decoder.ReadStringArray()
		if err != nil {
			logging.Error("Error decoding X-Live Subscribe Keys-2 %s %s", message, err)
			return
		}
	}
	pubsub.Subscribe(sqid, keys1, keys2)
}

// -----------------------------------------------------------------------------
// Logging
// -----------------------------------------------------------------------------

func logRequest(proto, status int, host string, request *http.Request) {
	var ip string
	splitPoint := strings.LastIndex(request.RemoteAddr, ":")
	if splitPoint == -1 {
		ip = request.RemoteAddr
	} else {
		ip = request.RemoteAddr[0:splitPoint]
	}
	logging.InfoData("fe", proto, status, request.Method, host, request.RawURL,
		ip, request.UserAgent(), request.Referer())
}

func filterRequestLog(record *logging.Record) (write bool, data []interface{}) {
	items := record.Items
	itemLength := len(items)
	if itemLength > 1 {
		identifier := items[0]
		switch identifier.(type) {
		case string:
			switch identifier.(string) {
			case "fe":
				return true, items[2 : itemLength-2]
			case "m":
				return true, items[1:itemLength]
			}
		}
	}
	return true, data
}

// -----------------------------------------------------------------------------
// Utility Functions
// -----------------------------------------------------------------------------

// The ``getErrorInfo`` utility function loads the specified error file from the
// given directory and returns its content and file size.
func getErrorInfo(directory, filename string) ([]byte, string) {
	path := filepath.Join(directory, filename)
	file, err := os.Open(path)
	if err != nil {
		runtime.StandardError(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		runtime.StandardError(err)
	}
	buffer := make([]byte, info.Size)
	_, err = file.Read(buffer[:])
	if err != nil && err != os.EOF {
		runtime.StandardError(err)
	}
	return buffer, fmt.Sprintf("%d", info.Size)
}

// The ``getFiles`` utility function populates the given ``mapping`` with
// ``StaticFile`` instances for all files found within a given ``directory``.
func getFiles(directory string, mapping map[string]*server.StaticFile, root string) {
	if debugMode {
		fmt.Printf("Caching static files in: %s\n", directory)
	}
	path, err := os.Open(directory)
	if err != nil {
		runtime.StandardError(err)
	}
	for {
		items, err := path.Readdir(100)
		if err != nil || len(items) == 0 {
			break
		}
		for _, item := range items {
			name := item.Name
			key := fmt.Sprintf("%s/%s", root, name)
			if item.IsDirectory() {
				getFiles(filepath.Join(directory, name), mapping, key)
			} else {
				content := make([]byte, item.Size)
				file, err := os.Open(filepath.Join(directory, name))
				if err != nil {
					runtime.StandardError(err)
				}
				_, err = file.Read(content[:])
				if err != nil && err != os.EOF {
					runtime.StandardError(err)
				}
				mimetype := mime.TypeByExtension(filepath.Ext(name))
				if mimetype == "" {
					mimetype = "application/octet-stream"
				}
				hash := sha1.New()
				hash.Write(content)
				buffer := &bytes.Buffer{}
				encoder := base64.NewEncoder(base64.URLEncoding, buffer)
				encoder.Write(hash.Sum())
				encoder.Close()
				mapping[key] = &server.StaticFile{
					content:  content,
					etag:     fmt.Sprintf(`"%s"`, buffer.String()),
					mimetype: mimetype,
					size:     fmt.Sprintf("%d", len(content)),
				}
			}
		}
	}
}

// The ``initFrontend`` utility function abstracts away the various checks and
// steps involved in setting up and running a new Frontend.
func initFrontend(status, host string, port int, officialHost, validAddress, cert, key, cometPrefix, websocketPrefix, instanceDirectory, upstreamHost string, upstreamPort int, upstreamTLS, maintenanceMode, liveMode bool, staticCache string, staticFiles map[string]*server.StaticFile, staticMaxAge int64) *Frontend {

	var err os.Error

	// Exit if the config values for the paths of the server's certificate or
	// key haven't been specified.
	if cert == "" {
		runtime.Error("ERROR: The %s-cert config value hasn't been specified.\n", status)
	}
	if key == "" {
		runtime.Error("ERROR: The %s-key config value hasn't been specified.\n", status)
	}

	// Initialise a fresh TLS Config.
	tlsConfig := &tls.Config{
		NextProtos: []string{"http/1.1"},
		Rand:       rand.Reader,
		Time:       time.Seconds,
	}

	// Load the certificate and private key into the TLS config.
	certPath := runtime.JoinPath(instanceDirectory, cert)
	keyPath := runtime.JoinPath(instanceDirectory, key)
	tlsConfig.Certificates = make([]tls.Certificate, 1)
	tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		runtime.Error("ERROR: Couldn't load %s certificate/key pair: %s\n",
			status, err)
	}

	// Instantiate the associated variables and listener for the HTTPS Frontend.
	upstreamAddr := fmt.Sprintf("%s:%d", upstreamHost, upstreamPort)
	frontendAddr := fmt.Sprintf("%s:%d", host, port)
	frontendConn, err := net.Listen("tcp", frontendAddr)
	if err != nil {
		runtime.Error("ERROR: Cannot listen on %s: %v\n", frontendAddr, err)
	}

	frontendListener := tls.NewListener(frontendConn, tlsConfig)

	// Compute the variables related to detecting valid hosts.
	var validWildcard bool
	if strings.HasPrefix(validAddress, "*.") {
		validAddress = validAddress[2:]
		validWildcard = true
	}

	// Compute the variables related to redirects.
	redirectURL := "https://" + officialHost
	redirectHTML := []byte(fmt.Sprintf(
		`Please <a href="%s">click here if your browser doesn't redirect</a> automatically.`,
		redirectURL))

	// Instantiate a ``Frontend`` object for use by the HTTPS Frontend.
	frontend := &Frontend{
		cometPrefix:     cometPrefix,
		liveMode:        liveMode,
		maintenanceMode: maintenanceMode,
		redirectHTML:    redirectHTML,
		redirectURL:     redirectURL,
		staticCache:     staticCache,
		staticFiles:     staticFiles,
		staticMaxAge:    staticMaxAge,
		upstreamAddr:    upstreamAddr,
		upstreamHost:    upstreamHost,
		upstreamTLS:     upstreamTLS,
		validAddress:    validAddress,
		validWildcard:   validWildcard,
		websocketPrefix: websocketPrefix,
	}

	// Start the HTTPS Frontend.
	go func() {
		err = http.Serve(frontendListener, frontend)
		if err != nil {
			runtime.Error("ERROR serving %s HTTPS Frontend: %s\n", status, err)
		}
	}()

	var frontendURL string
	if host == "" {
		frontendURL = fmt.Sprintf("https://localhost:%d", port)
	} else {
		frontendURL = fmt.Sprintf("https://%s:%d", host, port)
	}

	fmt.Printf("* HTTPS Frontend %s running on %s\n", status, frontendURL)

	return frontend

}

// -----------------------------------------------------------------------------
// Main Runner
// -----------------------------------------------------------------------------

func frontend(argv []string, usage string) {

	// Define the options for the command line and config file options parser.
	opts := optparse.Parser(
		"Usage: amp frontend <config.yaml> [options]\n\n    " + usage + "\n")

	httpsHost := opts.StringConfig("https-host", "",
		"the host to bind the HTTPS Frontends to")

	httpsPort := opts.IntConfig("https-port", 9040,
		"the base port for the HTTPS Frontends [9040]")

	officialHost := opts.StringConfig("offficial-host", "",
		"the official public host for the HTTPS Frontends")

	primaryHosts := opts.StringConfig("primary-hosts", "",
		"limit the primary HTTPS Frontend to the specified host pattern")

	primaryCert := opts.StringConfig("primary-cert", "cert/primary.cert",
		"the path to the primary host's TLS certificate [cert/primary.cert]")

	primaryKey := opts.StringConfig("primary-key", "cert/primary.key",
		"the path to the primary host's TLS key [cert/primary.key]")

	noSecondary := opts.BoolConfig("no-secondary", false,
		"disable the secondary HTTPS Frontend [false]")

	secondaryHosts := opts.StringConfig("secondary-hosts", "",
		"limit the secondary HTTPS Frontend to the specified host pattern")

	secondaryCert := opts.StringConfig("secondary-cert", "cert/secondary.cert",
		"the path to the secondary host's TLS certificate [cert/secondary.cert]")

	secondaryKey := opts.StringConfig("secondary-key", "cert/secondary.key",
		"the path to the secondary host's TLS key [cert/secondary.key]")

	errorDirectory := opts.StringConfig("error-dir", "error",
		"the path to the HTTP error files directory [error]")

	staticDirectory := opts.StringConfig("static-dir", "www",
		"the path to the static files directory [www]")

	staticMaxAge := opts.IntConfig("static-max-age", 86400,
		"max-age cache header value when serving the static files [86400]")

	noLivequery := opts.BoolConfig("no-livequery", false,
		"disable the LiveQuery node and WebSocket/Comet support [false]")

	websocketPrefix := opts.StringConfig("websocket-prefix", "/.ws/",
		"URL path prefix for WebSocket requests [/.ws/]")

	cometPrefix := opts.StringConfig("comet-prefix", "/.live/",
		"URL path prefix for Comet requests [/.live/]")

	livequeryHost := opts.StringConfig("livequery-host", "",
		"the host to bind the LiveQuery node to")

	livequeryPort := opts.IntConfig("livequery-port", 9050,
		"the port (both UDP and TCP) to bind the LiveQuery node to [9050]")

	livequeryExpiry := opts.IntConfig("livequery-expiry", 40,
		"maximum number of seconds a LiveQuery subscription is valid [40]")

	cookieKeyPath := opts.StringConfig("cookie-key", "cert/cookie.key",
		"the path to the file containing the key used to sign cookies [cert/cookie.key]")

	cookieName := opts.StringConfig("cookie-name", "user",
		"the property name of the cookie containing the user id [user]")

	acceptors := opts.StringConfig("acceptor-nodes", "localhost:9060",
		"comma-separated addresses of Acceptor nodes [localhost:9060]")

	acceptorKeyPath := opts.StringConfig("acceptor-key", "cert/acceptor.key",
		"the path to the file containing the Acceptor secret key [cert/acceptor.key]")

	acceptorIndex := opts.IntConfig("acceptor-index", 0,
		"this node's index in the Acceptor nodes address list [0]")

	leaseExpiry := opts.IntConfig("lease-expiry", 7,
		"maximum number of seconds a lease from an Acceptor node is valid [7]")

	noRedirect := opts.BoolConfig("no-redirect", false,
		"disable the HTTP Redirector [false]")

	httpHost := opts.StringConfig("http-host", "",
		"the host to bind the HTTP Redirector to")

	httpPort := opts.IntConfig("http-port", 9080,
		"the port to bind the HTTP Redirector to [9080]")

	redirectURL := opts.StringConfig("redirect-url", "",
		"the URL that the HTTP Redirector redirects to")

	pingPath := opts.StringConfig("ping-path", "/.ping",
		`URL path for a "ping" request [/.ping]`)
 
	enableHSTS := opts.BoolConfig("enable-hsts", false,
		"enable HTTP Strict Transport Security (HSTS) on redirects [false]")

	hstsMaxAge := opts.IntConfig("hsts-max-age", 50000000,
		"max-age value of HSTS in number of seconds [50000000]")

	upstreamHost := opts.StringConfig("upstream-host", "localhost",
		"the upstream host to connect to [localhost]")

	upstreamPort := opts.IntConfig("upstream-port", 8080,
		"the upstream port to connect to [8080]")

	upstreamTLS := opts.BoolConfig("upstream-tls", false,
		"use TLS when connecting to upstream [false]")

	maintenanceMode := opts.BoolConfig("maintenance", false,
		"start up in maintenance mode [false]")

	// Handle running as an Acceptor node if ``--run-as-acceptor`` was
	// specified.
	if *runAcceptor {

		// Exit if the `--acceptor-index`` is negative.
		if *acceptorIndex < 0 {
			runtime.Error("ERROR: The --acceptor-index cannot be negative.\n")
		}

		var index int
		var selfAddress string
		var acceptorNodes []string

		// Generate a list of all the acceptor node addresses and exit if we
		// couldn't find the address four ourselves at the given index.
		for _, acceptor := range strings.Split(*acceptors, ",") {
			acceptor = strings.TrimSpace(acceptor)
			if acceptor != "" {
				if index == *acceptorIndex {
					selfAddress = acceptor
				} else {
					acceptorNodes = append(acceptorNodes, acceptor)
				}
			}
			index += 1
		}

		if selfAddress == "" {
			runtime.Error("ERROR: Couldn't determine the address for the acceptor.\n")
		}

		// Initialise the process-related resources.
		runtime.InitProcess(fmt.Sprintf("acceptor-%d", *acceptorIndex), runPath)

		return

	}

	// Ensure that the directory containing static files exists.
	staticPath := runtime.JoinPath(instanceDirectory, *staticDirectory)
	dirInfo, err := os.Stat(staticPath)
	if err == nil {
		if !dirInfo.IsDirectory() {
			runtime.Error("ERROR: %q is not a directory\n", staticPath)
		}
	} else {
		runtime.StandardError(err)
	}

	// Load up all static files into a mapping.
	staticFiles := make(map[string]*server.StaticFile)
	getFiles(staticPath, staticFiles, "")

	// Pre-format the Cache-Control header for static files.
	staticCache := fmt.Sprintf("public, max-age=%d", *staticMaxAge)
	staticMaxAge64 := int64(*staticMaxAge)

	// Exit if the directory containing the 50x.html files isn't present.
	errorPath := runtime.JoinPath(instanceDirectory, *errorDirectory)
	dirInfo, err = os.Stat(errorPath)
	if err == nil {
		if !dirInfo.IsDirectory() {
			runtime.Error("ERROR: %q is not a directory\n", errorPath)
		}
	} else {
		runtime.StandardError(err)
	}

	// Load the content for the HTTP ``400``, ``500``, ``502`` and ``503``
	// errors.
	error400, error400Length = getErrorInfo(errorPath, "400.html")
	error500, error500Length = getErrorInfo(errorPath, "500.html")
	error502, error502Length = getErrorInfo(errorPath, "502.html")
	error503, error503Length = getErrorInfo(errorPath, "503.html")

	// Initialise the TLS config.
	tlsconf.Init()

	var liveMode bool

	// Setup the live support as long as it hasn't been disabled.
	if !*noLivequery {
		go handleLiveMessages()
		acceptorKey, err = ioutil.ReadFile(runtime.JoinPath(instanceDirectory, *acceptorKeyPath))
		if err != nil {
			runtime.StandardError(err)
		}
		cookieKey, err = ioutil.ReadFile(runtime.JoinPath(instanceDirectory, *cookieKeyPath))
		if err != nil {
			runtime.StandardError(err)
		}
		liveMode = true
		_ = *livequeryHost
		_ = *livequeryPort
		_ = *cookieName
		_ = *leaseExpiry
		livequeryTimeout = (int64(*livequeryExpiry) / 2) * 1000000000
	}

	// Create a container for the Frontend instances.
	frontends := make([]*Frontend, 0)

	// Create a channel which is used to toggle the state of the frontend's
	// maintenance mode based on process signals.
	maintenanceChannel := make(chan bool, 1)

	// Fork a goroutine which toggles the maintenance mode in a single place and
	// thus ensures thread safety.
	go func() {
		for {
			enabledState := <-maintenanceChannel
			for _, frontend := range frontends {
				if enabledState {
					frontend.maintenanceMode = true
				} else {
					frontend.maintenanceMode = false
				}
			}
		}
	}()

	// Register the signal handlers for SIGUSR1 and SIGUSR2.
	runtime.RegisterSignalHandler(os.SIGUSR1, func() {
		maintenanceChannel <- true
	})

	runtime.RegisterSignalHandler(os.SIGUSR2, func() {
		maintenanceChannel <- false
	})

	// Let the user know how many CPUs we're currently running on.
	fmt.Printf("Running the Amp Frontend on %d CPUs:\n", runtime.CPUCount)

	// If ``--public-address`` hasn't been specified, generate it from the given
	// frontend host and base port values -- assuming ``localhost`` for a blank
	// host.
	publicHost := *officialHost
	if publicHost == "" {
		if *httpsHost == "" {
			publicHost = fmt.Sprintf("localhost:%d", *httpsPort)
		} else {
			publicHost = fmt.Sprintf("%s:%d", *httpsHost, *httpsPort)
		}
	}

	// Setup and run the primary HTTPS Frontend.
	frontends = append(frontends, initFrontend("primary", *httpsHost,
		*httpsPort, publicHost, *primaryHosts, *primaryCert, *primaryKey,
		*cometPrefix, *websocketPrefix, instanceDirectory, *upstreamHost,
		*upstreamPort, *upstreamTLS, *maintenanceMode, liveMode, staticCache,
		staticFiles, staticMaxAge64))

	// Setup and run the secondary HTTPS Frontend.
	if !*noSecondary {
		frontends = append(frontends, initFrontend("secondary", *httpsHost,
			*httpsPort+1, publicHost, *secondaryHosts, *secondaryCert,
			*secondaryKey, *cometPrefix, *websocketPrefix, instanceDirectory,
			*upstreamHost, *upstreamPort, *upstreamTLS, *maintenanceMode,
			liveMode, staticCache, staticFiles, staticMaxAge64))
	}

	// Enter a wait loop if the HTTP Redirector has been disabled.
	if *noRedirect {
		loopForever := make(chan bool, 1)
		<-loopForever
	}

	// Otherwise, setup the HTTP Redirector.
	if *httpHost == "" {
		*httpHost = "localhost"
	}

	if *redirectURL == "" {
		*redirectURL = "https://" + publicHost
	}

	hsts := ""
	if *enableHSTS {
		hsts = fmt.Sprintf("max-age=%d", *hstsMaxAge)
	}

	httpAddr := fmt.Sprintf("%s:%d", *httpHost, *httpPort)
	httpListener, err := net.Listen("tcp", httpAddr)
	if err != nil {
		runtime.Error("ERROR: Cannot listen on %s: %v\n", httpAddr, err)
	}

	redirector := &server.HTTPRedirector{
		HSTS:       hsts,
		PingPath:   *pingPath,
		Pong:       []byte("pong"),
		PongLength: "4",
		URL:        *redirectURL,
	}

	// Start a goroutine which runs the HTTP redirector.
	go func() {
		err = http.Serve(httpListener, redirector)
		if err != nil {
			runtime.Error("ERROR serving HTTP Redirector: %s\n", err)
		}
	}()

	fmt.Printf("* HTTP Redirector running on http://%s:%d -> %s\n",
		*httpHost, *httpPort, *redirectURL)

	// Enter the wait loop for the process to be killed.
	loopForever := make(chan bool, 1)
	<-loopForever

}
