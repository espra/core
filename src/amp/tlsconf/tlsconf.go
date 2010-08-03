// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// The tls package provides utilities to support TLS connections. When the
// package is initialised via ``tlsconf.Init()``, it generates a default
// configuration from the TLS Certificate Authorities data included with Ampify.
package tlsconf

import (
	"amp/runtime"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

var Config *tls.Config

func GenConfig(file string) (config *tls.Config, err os.Error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	roots := tls.NewCASet()
	roots.SetFromPEM(data)
	config = &tls.Config{
		Rand:    rand.Reader,
		Time:    time.Seconds,
		RootCAs: roots,
	}
	return config, nil
}

// Set the ``tlsconf.Config`` variable.
func Init() {
	path := runtime.AmpifyRoot + "/environ/local/var/ca.cert"
	var err os.Error
	Config, err = GenConfig(path)
	if err != nil {
		fmt.Printf("ERROR: Couldn't load %s: %s\n", path, err)
		os.Exit(1)
	}
}
