// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package server

// Constants for the different log event types.
const (
	HTTP_PING = iota
	HTTP_REDIRECT
	HTTPS_COMET
	HTTPS_INTERNAL_ERROR
	HTTPS_MAINTENANCE
	HTTPS_PROXY_ERROR
	HTTPS_REDIRECT
	HTTPS_STATIC
	HTTPS_UPSTREAM
	HTTPS_WEBSOCKET
)
