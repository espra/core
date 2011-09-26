// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// HTTPS Frontend
// ==============
//
package server

import (
	"amp/logging"
	"amp/tlsconf"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"http"
	"io"
	"io/ioutil"
	"json"
	"net"
	"strconv"
	"strings"
	"time"
	"websocket"
)

// The ``StaticFile`` type holds the data needed to serve a static file via the
// HTTPS Frontend.
type StaticFile struct {
	Content  []byte
	ETag     string
	Mimetype string
	Size     string
}

// The type for the HTTPS Frontend.
type Frontend struct {

	// Logging function.
	Log func(int, int, string, *http.Request)

	// Error pages.
	Error400       []byte
	Error500       []byte
	Error502       []byte
	Error503       []byte
	Error400Length string
	Error500Length string
	Error502Length string
	Error503Length string

	// Live support.
	LiveMode        bool
	CometPrefix     string
	WebsocketPrefix string

	// Debug mode.
	Debug bool

	// Maintenance mode toggle.
	MaintenanceMode bool

	// Redirects.
	RedirectHTML []byte
	RedirectURL  string

	// Static files.
	StaticCache  string
	StaticFiles  map[string]*StaticFile
	StaticMaxAge int64

	upstreamAddr string
	upstreamHost string
	upstreamTLS  bool

	// Host validation.
	ValidAddress  string
	ValidWildcard bool
}

func (frontend *Frontend) ServeHTTP(conn http.ResponseWriter, req *http.Request) {

	originalHost := req.Host

	// Redirect all requests to the "official" public host if the Host header
	// doesn't match.
	if !frontend.isValidHost(originalHost) {
		conn.Header().Set("Location", frontend.RedirectURL)
		conn.WriteHeader(http.StatusMovedPermanently)
		conn.Write(frontend.RedirectHTML)
		frontend.Log(HTTPS_REDIRECT, http.StatusMovedPermanently, originalHost, req)
		return
	}

	// Return the HTTP 503 error page if we're in maintenance mode.
	if frontend.MaintenanceMode {
		headers := conn.Header()
		headers.Set("Content-Type", "text/html; charset=utf-8")
		headers.Set("Content-Length", frontend.Error503Length)
		conn.WriteHeader(http.StatusServiceUnavailable)
		conn.Write(frontend.Error503)
		frontend.Log(HTTPS_MAINTENANCE, http.StatusServiceUnavailable, originalHost, req)
		return
	}

	reqPath := req.URL.Path

	// Handle requests for any files exposed within the static directory.
	if staticFile, ok := frontend.StaticFiles[reqPath]; ok {
		expires := time.SecondsToUTC(time.Seconds() + frontend.StaticMaxAge)
		headers := conn.Header()
		headers.Set("Expires", expires.Format(http.TimeFormat))
		headers.Set("Cache-Control", frontend.StaticCache)
		headers.Set("Etag", staticFile.ETag)
		if req.Header.Get("If-None-Match") == staticFile.ETag {
			conn.WriteHeader(http.StatusNotModified)
			frontend.Log(HTTPS_STATIC, http.StatusNotModified, originalHost, req)
			return
		}
		// Special case /.well-known/oauth.json?callback= requests.
		if reqPath == "/.well-known/oauth.json" && req.URL.RawQuery != "" {
			query, err := http.ParseQuery(req.URL.RawQuery)
			if err != nil {
				logging.Error("Error parsing oauth.json query string %q: %s",
					req.URL.RawQuery, err)
				frontend.ServeError400(conn, originalHost, req)
				return
			}
			if callbackList, found := query["callback"]; found {
				callback := callbackList[0]
				if callback != "" {
					respLen := len(callback) + len(staticFile.Content) + 2
					headers.Set("Content-Type", "text/javascript")
					headers.Set("Content-Length", fmt.Sprintf("%d", respLen))
					conn.WriteHeader(http.StatusOK)
					conn.Write([]byte(callback))
					conn.Write([]byte{'('})
					conn.Write(staticFile.Content)
					conn.Write([]byte{')'})
					frontend.Log(HTTPS_STATIC, http.StatusOK, originalHost, req)
					return
				}
			}
		}
		headers.Set("Content-Type", staticFile.Mimetype)
		headers.Set("Content-Length", staticFile.Size)
		conn.WriteHeader(http.StatusOK)
		conn.Write(staticFile.Content)
		frontend.Log(HTTPS_STATIC, http.StatusOK, originalHost, req)
		return
	}

	if frontend.LiveMode {

		// Handle WebSocket requests.
		if strings.HasPrefix(reqPath, frontend.WebsocketPrefix) {
			websocket.Handler(frontend.getWebSocketHandler()).ServeHTTP(conn, req)
			return
		}

		// Handle long-polling Comet requests.
		if strings.HasPrefix(reqPath, frontend.CometPrefix) {
			query, err := http.ParseQuery(req.URL.RawQuery)
			if err != nil {
				logging.Error("Error parsing Comet query string %q: %s",
					req.URL.RawQuery, err)
				frontend.ServeError400(conn, originalHost, req)
				return
			}
			queryReq, found := query["q"]
			if !found {
				frontend.ServeError400(conn, originalHost, req)
				return
			}
			response, status := getLiveItems(queryReq[0])
			headers := conn.Header()
			headers.Set("Content-Type", "application/json")
			headers.Set("Content-Length", fmt.Sprintf("%d", len(response)))
			conn.WriteHeader(status)
			conn.Write(response)
			frontend.Log(HTTPS_COMET, status, originalHost, req)
			return
		}

	}

	// Open a connection to the upstream server.
	upstreamConn, err := net.Dial("tcp", frontend.upstreamAddr)
	if err != nil {
		logging.Error("Couldn't connect to upstream: %s", err)
		frontend.ServeError502(conn, originalHost, req)
		return
	}

	var clientIP string
	var upstream net.Conn

	splitPoint := strings.LastIndex(req.RemoteAddr, ":")
	if splitPoint == -1 {
		clientIP = req.RemoteAddr
	} else {
		clientIP = req.RemoteAddr[0:splitPoint]
	}

	if frontend.upstreamTLS {
		upstream = tls.Client(upstreamConn, tlsconf.Config)
		defer upstream.Close()
	} else {
		upstream = upstreamConn
	}

	// Modify the request Host: and User-Agent: headers.
	req.Host = frontend.upstreamHost
	req.Header.Set(
		"User-Agent",
		fmt.Sprintf("%s, %s, %s", req.UserAgent(), clientIP, originalHost))

	// Send the request to the upstream server.
	err = req.Write(upstream)
	if err != nil {
		logging.Error("Error writing to the upstream server: %s", err)
		frontend.ServeError502(conn, originalHost, req)
		return
	}

	// Parse the response from upstream.
	resp, err := http.ReadResponse(bufio.NewReader(upstream), req)
	if err != nil {
		logging.Error("Error parsing response from upstream: %s", err)
		frontend.ServeError502(conn, originalHost, req)
		return
	}

	defer resp.Body.Close()

	// Get the original request header.
	headers := conn.Header()

	// Set a variable to hold the X-Live header value if present.
	var liveLength int

	if frontend.LiveMode {
		xLive := resp.Header.Get("X-Live")
		if xLive != "" {
			// If the X-Live header was set, parse it into an int.
			liveLength, err = strconv.Atoi(xLive)
			if err != nil {
				logging.Error("Error converting X-Live header value %q: %s", xLive, err)
				frontend.ServeError500(conn, originalHost, req)
				return
			}
			resp.Header.Del("X-Live")
		}
	}

	var body []byte

	if liveLength > 0 {

		var gzipSet bool
		var respBody io.ReadCloser

		// Check Content-Encoding to see if upstream sent gzipped content.
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzipSet = true
			respBody, err = gzip.NewReader(resp.Body)
			if err != nil {
				logging.Error("Error reading gzipped response from upstream: %s", err)
				frontend.ServeError500(conn, originalHost, req)
				return
			}
			defer respBody.Close()
		} else {
			respBody = resp.Body
		}

		// Read the X-Live content from the response body.
		liveMessage := make([]byte, liveLength)
		n, err := respBody.Read(liveMessage)
		if n != liveLength || err != nil {
			logging.Error("Error reading X-Live response from upstream: %s", err)
			frontend.ServeError500(conn, originalHost, req)
			return
		}

		// Read the response to send back to the original request.
		body, err = ioutil.ReadAll(respBody)
		if err != nil {
			logging.Error("Error reading non X-Live response from upstream: %s", err)
			frontend.ServeError500(conn, originalHost, req)
			return
		}

		// Re-encode the response if it had been gzipped by upstream.
		if gzipSet {
			buffer := &bytes.Buffer{}
			encoder, err := gzip.NewWriter(buffer)
			if err != nil {
				logging.Error("Error creating a new gzip Writer: %s", err)
				frontend.ServeError500(conn, originalHost, req)
				return
			}
			n, err = encoder.Write(body)
			if n != len(body) || err != nil {
				logging.Error("Error writing to the gzip Writer: %s", err)
				frontend.ServeError500(conn, originalHost, req)
				return
			}
			err = encoder.Close()
			if err != nil {
				logging.Error("Error finalising the write to the gzip Writer: %s", err)
				frontend.ServeError500(conn, originalHost, req)
				return
			}
			body = buffer.Bytes()
		}

		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
		liveChannel <- liveMessage

	} else {
		// Read the full response body.
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			logging.Error("Error reading response from upstream: %s", err)
			frontend.ServeError502(conn, originalHost, req)
			return
		}
	}

	// Set the received headers back to the initial connection.
	for k, values := range resp.Header {
		for _, v := range values {
			headers.Add(k, v)
		}
	}

	// Write the response body back to the initial connection.
	conn.WriteHeader(resp.StatusCode)
	conn.Write(body)
	frontend.Log(HTTPS_UPSTREAM, resp.StatusCode, originalHost, req)

}

// -----------------------------------------------------------------------------
// WebSocket Handler
// -----------------------------------------------------------------------------

func (frontend *Frontend) getWebSocketHandler() func(*websocket.Conn) {
	return func(conn *websocket.Conn) {
		respStatus := http.StatusOK
		defer func() {
			conn.Close()
			frontend.Log(HTTPS_WEBSOCKET, respStatus, conn.Request.Host, conn.Request)
		}()
		request := make([]byte, 1024)
		for {
			n, err := conn.Read(request)
			if err != nil {
				if frontend.Debug {
					logging.Error("Error reading on WebSocket: %s", err)
				}
				break
			}
			response, status := getLiveItems(string(request[:n]))
			if status != http.StatusOK {
				respStatus = status
				break
			}
			if len(response) != 0 {
				if _, err = conn.Write(response); err != nil {
					break
				}
			}
		}
	}
}

// -----------------------------------------------------------------------------
// Host Validator
// -----------------------------------------------------------------------------

// The ``isValidHost`` method validates a request Host against any specified
// valid address/wildcard.
func (frontend *Frontend) isValidHost(host string) (valid bool) {
	if frontend.ValidAddress == "" {
		return true
	}
	if frontend.ValidWildcard {
		splitHost := strings.SplitN(host, ".", 2)
		if len(splitHost) != 2 {
			return
		}
		if splitHost[1] == frontend.ValidAddress {
			return true
		}
		return
	}
	if host == frontend.ValidAddress {
		return true
	}
	return
}

// -----------------------------------------------------------------------------
// Error Handlers
// -----------------------------------------------------------------------------

func (frontend *Frontend) ServeError400(conn http.ResponseWriter, host string, request *http.Request) {
	headers := conn.Header()
	headers.Set("Content-Type", "text/html; charset=utf-8")
	headers.Set("Content-Length", frontend.Error400Length)
	conn.WriteHeader(http.StatusBadRequest)
	conn.Write(frontend.Error400)
	frontend.Log(HTTPS_PROXY_ERROR, http.StatusBadRequest, host, request)
}

func (frontend *Frontend) ServeError500(conn http.ResponseWriter, host string, request *http.Request) {
	headers := conn.Header()
	headers.Set("Content-Type", "text/html; charset=utf-8")
	headers.Set("Content-Length", frontend.Error500Length)
	conn.WriteHeader(http.StatusInternalServerError)
	conn.Write(frontend.Error500)
	frontend.Log(HTTPS_INTERNAL_ERROR, http.StatusInternalServerError, host, request)
}

func (frontend *Frontend) ServeError502(conn http.ResponseWriter, host string, request *http.Request) {
	headers := conn.Header()
	headers.Set("Content-Type", "text/html; charset=utf-8")
	headers.Set("Content-Length", frontend.Error502Length)
	conn.WriteHeader(http.StatusBadGateway)
	conn.Write(frontend.Error502)
	frontend.Log(HTTPS_PROXY_ERROR, http.StatusBadGateway, host, request)
}

// -----------------------------------------------------------------------------
// Live Listener
// -----------------------------------------------------------------------------

func getLiveItems(request string) ([]byte, int) {
	reqs := strings.Split(request, ",")
	if len(reqs) == 0 {
		return nil, http.StatusBadRequest
	}
	sid := reqs[0] /* XXX validate sid */
	result, refresh, ok := pubsub.Listen(sid, reqs[1:], livequeryTimeout)
	if ok {
		data := make(map[string]map[string][]string)
		data["items"] = result
		response, err := json.Marshal(data)
		if err != nil {
			logging.Error("ERROR encoding JSON: %s for: %v", err, data)
			return nil, http.StatusInternalServerError
		}
		return response, http.StatusOK
	}
	refreshCount := len(refresh)
	if refreshCount != 0 {
		data := make(map[string][]string)
		qids := make([]string, refreshCount)
		i := 0
		for qid, _ := range refresh {
			qids[i] = qid
			i++
		}
		data["refresh"] = qids
		response, err := json.Marshal(data)
		if err != nil {
			logging.Error("ERROR encoding JSON: %s for: %v", err, data)
			return nil, http.StatusInternalServerError
		}
		return response, http.StatusOK
	}
	return nil, http.StatusNotFound
}
