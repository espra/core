// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// Amp Frontend
// ============
//
package main

import (
	"amp/log"
	"amp/master"
	"amp/optparse"
	"amp/runtime"
	"amp/server"
	"amp/tlsconf"
	"fmt"
	"os"
	"strings"
)

func getValidAddr(hosts string) (bool, string) {
	var wildcard bool
	if strings.HasPrefix(hosts, "*.") {
		hosts = hosts[2:]
		wildcard = true
	}
	return wildcard, hosts
}

func ampFrontend(argv []string, usage string) {

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

	hstsMaxAge := opts.IntConfig("hsts-max-age", 50000000,
		"max-age in seconds for HTTP Strict Transport Security [50000000]")

	noRedirect := opts.BoolConfig("no-redirect", false,
		"disable the HTTP Redirector [false]")

	httpHost := opts.StringConfig("http-host", "",
		"the host to bind the HTTP Redirector to")

	httpPort := opts.IntConfig("http-port", 9080,
		"the port to bind the HTTP Redirector to [9080]")

	httpRedirectURL := opts.StringConfig("redirect-url", "",
		"the URL that the HTTP Redirector redirects to")

	singleNode := opts.StringConfig("single-node", "",
		"the upstream single node address if running without a master")

	masterNodes := opts.StringConfig("master-nodes", "localhost:8060",
		"comma-separated addresses of amp master nodes [localhost:8060]")

	masterCert := opts.StringConfig("master-cert", "cert/master.cert",
		"the path to the amp master certificate [cert/master.cert]")

	ironKeyPath := opts.StringConfig("iron-key", "cert/iron.key",
		"the path to the key used for iron strings [cert/iron.key]")

	maintenanceMode := opts.BoolConfig("maintenance", false,
		"start up in maintenance mode [false]")

	_, instanceDirectory, _ := runtime.DefaultOpts("frontend", opts, argv)

	// Ensure that the directory containing static files exists.
	staticPath := runtime.JoinPath(instanceDirectory, *staticDirectory)
	dirInfo, err := os.Stat(staticPath)
	if err == nil {
		if !dirInfo.IsDirectory() {
			runtime.Error("%q is not a directory", staticPath)
		}
	} else {
		runtime.StandardError(err)
	}

	// Ensure that the directory containing error files exists.
	errorPath := runtime.JoinPath(instanceDirectory, *errorDirectory)
	dirInfo, err = os.Stat(errorPath)
	if err == nil {
		if !dirInfo.IsDirectory() {
			runtime.Error("%q is not a directory", errorPath)
		}
	} else {
		runtime.StandardError(err)
	}

	// If ``--official-host`` hasn't been specified, generate it from the given
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

	// Compute the HSTS max age header value.
	hsts := fmt.Sprintf("max-age=%d", *hstsMaxAge)

	// Pre-format the Cache-Control header for static files.
	staticCache := fmt.Sprintf("public, max-age=%d", *staticMaxAge)
	staticMaxAge64 := int64(*staticMaxAge)

	// Compute the variables related to redirects.
	redirectURL := "https://" + publicHost
	redirectHTML := []byte(fmt.Sprintf(
		`Please <a href="%s">click here if your browser doesn't redirect</a> automatically.`,
		redirectURL))

	// Compute the path to the Iron key.
	ironPath := runtime.JoinPath(instanceDirectory, *ironKeyPath)

	// Instantiate the master client.
	masterClient, err := master.NewClient(
		*masterNodes, runtime.JoinPath(instanceDirectory, *masterCert))

	if err != nil {
		runtime.StandardError(err)
	}

	var noMaster bool
	if *singleNode != "" {
		noMaster = true
	}

	// Let the user know how many CPUs we're currently running on.
	log.Info("Running the Amp Frontend on %d CPUs.", runtime.CPUCount)

	// Initialise the TLS config.
	tlsconf.Init()

	// Initialise a container for the HTTPSFrontends.
	webFrontends := make([]*server.HTTPSFrontend, 1)

	// Compute the variables related to detecting valid hosts.
	primaryWildcard, primaryAddr := getValidAddr(*primaryHosts)
	secondaryWildcard, secondaryAddr := getValidAddr(*secondaryHosts)

	// Instantiate the primary ``HTTPSFrontend`` object.
	frontend := &server.HTTPSFrontend{
		HSTS:            hsts,
		MaintenanceMode: *maintenanceMode,
		MasterClient:    masterClient,
		NoMaster:        noMaster,
		RedirectHTML:    redirectHTML,
		RedirectURL:     redirectURL,
		SingleNode:      *singleNode,
		StaticCache:     staticCache,
		StaticMaxAge:    staticMaxAge64,
		ValidAddress:    primaryAddr,
		ValidWildcard:   primaryWildcard,
	}

	frontend.LoadAssets(errorPath, ironPath, staticPath)
	frontend.Run(*httpsHost, *httpsPort,
		runtime.JoinPath(instanceDirectory, *primaryCert),
		runtime.JoinPath(instanceDirectory, *primaryKey))

	webFrontends[0] = frontend

	// Setup and run the secondary HTTPSFrontend.
	if !*noSecondary {
		frontend = &server.HTTPSFrontend{
			HSTS:            hsts,
			MaintenanceMode: *maintenanceMode,
			MasterClient:    masterClient,
			NoMaster:        noMaster,
			RedirectHTML:    redirectHTML,
			RedirectURL:     redirectURL,
			SingleNode:      *singleNode,
			StaticCache:     staticCache,
			StaticMaxAge:    staticMaxAge64,
			ValidAddress:    secondaryAddr,
			ValidWildcard:   secondaryWildcard,
		}
		frontend.LoadAssets(errorPath, ironPath, staticPath)
		frontend.Run(*httpsHost, *httpsPort+1,
			runtime.JoinPath(instanceDirectory, *secondaryCert),
			runtime.JoinPath(instanceDirectory, *secondaryKey))
		webFrontends = append(webFrontends, frontend)
	}

	// Create a channel which is used to toggle maintenance mode based on
	// process signals.
	maintenanceChannel := make(chan bool, 1)

	// Fork a goroutine which toggles the maintenance mode in a single place and
	// thus ensures thread safety.
	go func() {
		for {
			enabledState := <-maintenanceChannel
			for _, frontend := range webFrontends {
				if enabledState {
					frontend.MaintenanceMode = true
				} else {
					frontend.LoadAssets(errorPath, ironPath, staticPath)
					frontend.MaintenanceMode = false
				}
			}
		}
	}()

	// Register the signal handlers for SIGUSR1 and SIGUSR2.
	runtime.SignalHandlers[os.SIGUSR1] = func() {
		maintenanceChannel <- true
	}

	runtime.SignalHandlers[os.SIGUSR2] = func() {
		maintenanceChannel <- false
	}

	// Enter a wait loop if the HTTP Redirector has been disabled.
	if *noRedirect {
		loopForever := make(chan bool, 1)
		<-loopForever
	}

	// Otherwise, setup and run the HTTP Redirector.
	if *httpHost == "" {
		*httpHost = "localhost"
	}

	if *httpRedirectURL == "" {
		*httpRedirectURL = "https://" + publicHost
	}

	redirector := &server.HTTPRedirector{*httpRedirectURL}
	redirector.Run(*httpHost, *httpPort)

	// Enter the wait loop for the process to be killed.
	loopForever := make(chan bool, 1)
	<-loopForever

}
