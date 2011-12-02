// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// HTTP Redirector
// ===============
//
package server

import (
	"fmt"
	"net/http"
)

type HTTPRedirector struct {
	HSTS       string
	Message    string
	PingPath   string
	Pong       []byte
	PongLength string
	URL        string
}

func (redirector *HTTPRedirector) ServeHTTP(conn http.ResponseWriter, req *http.Request) {

	if req.URL.Path == redirector.PingPath {
		conn.Header().Set("Content-Type", "text/plain")
		conn.Header().Set("Content-Length", redirector.PongLength)
		conn.WriteHeader(http.StatusOK)
		conn.Write(redirector.Pong)
		logWebRequest(HTTP_PING, req)
		return
	}

	var url string
	if len(req.URL.RawQuery) > 0 {
		url = fmt.Sprintf("%s%s?%s", redirector.URL, req.URL.Path, req.URL.RawQuery)
	} else {
		url = fmt.Sprintf("%s%s", redirector.URL, req.URL.Path)
	}

	if len(url) == 0 {
		url = "/"
	}

	conn.Header().Set("Location", url)
	conn.Header().Set("Strict-Transport-Security", redirector.HSTS)
	conn.WriteHeader(http.StatusMovedPermanently)
	fmt.Fprintf(conn, redirector.Message, url)
	logWebRequest(HTTP_REDIRECT, req)

}
