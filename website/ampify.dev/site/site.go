// Public Domain (-) 2018-present, The Amp Authors.
// See the Amp UNLICENSE file for details.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var goMeta = []byte(`<!doctype html>
<meta name="go-import" content="ampify.dev git https://github.com/ampify/amp">
<meta name="go-source" content="ampify.dev https://github.com/ampify/amp https://github.com/ampify/amp/tree/master{/dir} https://github.com/ampify/amp/blob/master{/dir}/{file}#L{line}">`)

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
		case "cmd", "pkg":
			if r.URL.RawQuery != "" {
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, "https://godoc.org/ampify.dev/"+strings.Join(split[1:], "/"), http.StatusFound)
			return
		}
	}
	http.Redirect(w, r, "https://github.com/ampify/amp", http.StatusFound)
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
