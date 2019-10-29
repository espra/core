// Public Domain (-) 2018-present, The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const slackURL = "https://join.slack.com/t/ampifyhq/shared_invite/enQtNzAzNzMzNDExNjAzLWM2YmE3MjFkNGQ3OTgyM2U4Yjc2ZjIyMjI1MjE1YTk1MjIxNzRhMTMzMzVjNWJhZTUwOGJiMGM1ZTFlZGI5OGQ"

var goMeta = []byte(`<!doctype html>
<meta name="go-import" content="ampify.dev git https://github.com/espra/ampify">
<meta name="go-source" content="ampify.dev https://github.com/espra/ampify https://github.com/espra/ampify/tree/master{/dir} https://github.com/espra/ampify/blob/master{/dir}/{file}#L{line}">`)

func handle(w http.ResponseWriter, r *http.Request) {
	if isAppEngine() {
		if r.Host != "ampify.dev" || r.Header.Get("X-Forwarded-Proto") == "http" {
			url := r.URL
			url.Host = "ampify.dev"
			url.Scheme = "https"
			http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
			return
		}
	}
	if r.URL.Query().Get("go-get") != "" {
		w.Write(goMeta)
		return
	}
	split := strings.Split(r.URL.Path, "/")
	if len(split) >= 2 {
		switch split[1] {
		case "cmd", "go":
			if r.URL.RawQuery != "" {
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, "https://godoc.org/ampify.dev/"+strings.Join(split[1:], "/"), http.StatusFound)
			return
		case "slack":
			http.Redirect(w, r, slackURL, http.StatusFound)
		}
	}
	http.Redirect(w, r, "https://github.com/espra/ampify", http.StatusFound)
}

func isAppEngine() bool {
	return os.Getenv("GAE_INSTANCE") != ""
}

func main() {
	http.HandleFunc("/", handle)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
