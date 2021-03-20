// Public Domain (-) 2018-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

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
	goGetResponse = `<!doctype html>
<meta name="go-import" content="web4.cc git https://github.com/espra/web4">
<meta name="go-source" content="web4.cc https://github.com/espra/web4 https://github.com/espra/web4/tree/main{/dir} https://github.com/espra/web4/blob/main{/dir}/{file}#L{line}">`
	slackInviteURL = "https://join.slack.com/t/espra/shared_invite/enQtOTAxMjMxOTI3NDkwLTc4YWUxYzIxMmNjMDU3MmVhNjA2YTc4YjUxZDQwNjgzZTcxMmJiMDU2YmQyNDdmMmUxZTM2OWU0ODUyMGJkODY"
)

// We assume a prod environment if the PORT environment variable has been set.
var isProd bool

func handle(w http.ResponseWriter, r *http.Request) {
	if isProd {
		if r.Host != "web4.cc" || r.Header.Get("X-Forwarded-Proto") == "http" {
			url := r.URL
			url.Host = "web4.cc"
			url.Scheme = "https"
			http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
			return
		}
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	}
	if r.URL.Query().Get("go-get") != "" {
		w.Write([]byte(goGetResponse))
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
			http.Redirect(w, r, "https://pkg.go.dev/web4.cc/"+strings.Join(split[1:], "/"), http.StatusFound)
			return
		case "health":
			w.Write([]byte("OK"))
			return
		case "slack":
			http.Redirect(w, r, slackInviteURL, http.StatusFound)
			return
		}
	}
	http.Redirect(w, r, "https://github.com/espra/web4", http.StatusFound)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	} else {
		isProd = true
	}
	http.HandleFunc("/", handle)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      http.DefaultServeMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("Listening on http://localhost%s\n", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to run HTTP server: %s\n", err)
	}
}
