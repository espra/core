// Public Domain (-) 2018-present, The Core Authors.
// See the Core UNLICENSE file for details.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	goGetHeader = `<!doctype html>
<meta name="go-import" content="dappui.com git https://github.com/dappui/core">
<meta name="go-source" content="dappui.com https://github.com/dappui/core https://github.com/dappui/core/tree/master{/dir} https://github.com/dappui/core/blob/master{/dir}/{file}#L{line}">`
	slackInviteURL = "https://join.slack.com/t/dappui/shared_invite/enQtOTAxMjMxOTI3NDkwLTc4YWUxYzIxMmNjMDU3MmVhNjA2YTc4YjUxZDQwNjgzZTcxMmJiMDU2YmQyNDdmMmUxZTM2OWU0ODUyMGJkODY"
)

// We assume a prod environment if the PORT environment variable has been set.
var isProd bool

func handle(w http.ResponseWriter, r *http.Request) {
	if isProd {
		if r.Host != "dappui.com" || r.Header.Get("X-Forwarded-Proto") == "http" {
			url := r.URL
			url.Host = "dappui.com"
			url.Scheme = "https"
			http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
			return
		}
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	}
	if r.URL.Query().Get("go-get") != "" {
		w.Write([]byte(goGetHeader))
		return
	}
	split := strings.Split(r.URL.Path, "/")
	if len(split) >= 2 {
		switch split[1] {
		case "cmd", "infra", "pkg", "service", "website":
			if r.URL.RawQuery != "" {
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, "https://godoc.org/dappui.com/"+strings.Join(split[1:], "/"), http.StatusFound)
			return
		case "health":
			w.Write([]byte("OK"))
			return
		case "slack":
			http.Redirect(w, r, slackInviteURL, http.StatusFound)
			return
		}
	}
	http.Redirect(w, r, "https://github.com/dappui/core", http.StatusFound)
}

func main() {
	isProd = os.Getenv("PRODUCTION") == "1"
	http.HandleFunc("/", handle)
	go func() {
		if isProd {
			return
		}
		srv := &http.Server{
			Addr:         ":8080",
			Handler:      http.DefaultServeMux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		log.Printf("Listening on http://localhost%s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to run HTTP server: %s", err)
		}
	}()
	go func() {
		port := 8443
		if isProd {
			port = 443
		}
		srv := &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      http.DefaultServeMux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		log.Printf("Listening on https://localhost%s\n", srv.Addr)
		if err := srv.ListenAndServeTLS("selfsigned/tls.cert", "selfsigned/tls.key"); err != nil {
			log.Fatalf("Failed to run HTTPS server: %s", err)
		}
	}()
	select {}
}
