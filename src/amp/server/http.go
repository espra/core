// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// HTTP Redirector
// ===============
//
package server

import (
	"http"
	"fmt"
)

type HTTPRedirector struct {
	Log        func(int, int, string, *http.Request)
	Message    string
	HSTS       string
	PingPath   string
	Pong       []byte
	PongLength string
	URL        string
}

func (redirector *HTTPRedirector) ServeHTTP(conn http.ResponseWriter, req *http.Request) {

	if req.URL.Path == redirector.PingPath {
		headers := conn.Header()
		headers.Set("Content-Type", "text/plain")
		headers.Set("Content-Length", redirector.PongLength)
		conn.WriteHeader(http.StatusOK)
		conn.Write(redirector.Pong)
		redirector.Log(HTTP_PING, http.StatusOK, req.Host, req)
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

	if redirector.HSTS != "" {
		conn.Header().Set("Strict-Transport-Security", redirector.HSTS)
	}

	conn.Header().Set("Location", url)
	conn.WriteHeader(http.StatusMovedPermanently)
	fmt.Fprintf(conn, redirector.Message, url)
	redirector.Log(HTTP_REDIRECT, http.StatusMovedPermanently, req.Host, req)

}
