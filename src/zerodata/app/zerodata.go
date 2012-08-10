// Public Domain (-) 2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package zerodata

import (
	"appengine"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	// Provide some basic info for GET requests.
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, "Zerodata API Endpoint.")
		return
	}
	// Instantiate the App Engine Context object.
	ctx := appengine.NewContext(r)
	// Do some sanity checks when in production.
	if !appengine.IsDevAppServer() {
		// Ensure that we are only being accessed over HTTPS.
		if r.TLS == nil {
			raise(400, ctx, w, "Endpoint must only be accessed over HTTPS in production.")
			return
		}
		// Check the IP address against the whitelist.
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			raise(400, ctx, w, "Couldn't parse the remote address: %s", err)
			return
		}
		if !isWhitelisted(ctx, host) {
			raise(400, ctx, w, "Sorry, the IP address %q hasn't been whitelisted.", host)
			return
		}
	}
	// Check the secret key.
	key := make([]byte, secretLength)
	n, err := r.Body.Read(key)
	if err != nil || n != secretLength {
		raise(401, ctx, w, "Sorry, invalid secret key.")
		return
	}
	if subtle.ConstantTimeCompare(key, secret) != 1 {
		raise(401, ctx, w, "Sorry, invalid secret key.")
		return
	}
	if len(r.URL.Path) == 0 {
		http.NotFound(w, r)
		return
	}
	// Delegate the request to the RPC handler.
	handleRPC(ctx, w, r)
}

func raise(status int, ctx appengine.Context, w http.ResponseWriter, s string, args ...interface{}) {
	ctx.Errorf(s, args...)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprint(w, "ERROR: ")
	fmt.Fprintf(w, s, args...)
}

func init() {
	if secretLength == 0 {
		panic("the secret key cannot be empty.")
	}
	if secretLength != len(secret) {
		panic("the secret key length does not match the `secretLength` variable.")
	}
	http.HandleFunc("/", handler)
}
