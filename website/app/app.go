// Public Domain (-) 2018-present, The Core Authors.
// See the Core UNLICENSE file for details.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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
		case "slack":
			http.Redirect(w, r, slackInviteURL, http.StatusFound)
			return
		}
	}
	http.Redirect(w, r, "https://github.com/dappui/core", http.StatusFound)
}

func main() {
	http.HandleFunc("/", handle)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	} else {
		isProd = true
	}
	log.Printf("Listening on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
