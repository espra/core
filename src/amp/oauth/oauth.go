// Public Domain (-) 2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// Based on http://code.google.com/o/goauth2 but increasingly divergent

package oauth

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// OAuthService: an OAuth 2 consumer service.
type OAuthService struct {
	ClientID     string
	ClientSecret string
	Scope        string
	AuthURL      string
	TokenURL     string
	RedirectURL  string // Defaults to out-of-band mode if empty.
	AcceptHeader string
}

func (o *OAuthService) redirectURL() string {
	if o.RedirectURL != "" {
		return o.RedirectURL
	}
	return "oob"
}

// Token contains an end-user's tokens.
// This is the data you must store to persist authentication.
type Token struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time // If zero the token has no (known) expiry time.
}

// Checks if a Token is expired
func (t *Token) Expired() bool {
	if t.Expiry.IsZero() {
		return false
	}
	return t.Expiry.Before(time.Now())
}

// AuthCodeURL returns a URL that the end-user should be redirected to,
// so that they may obtain an authorization code.
func (c *OAuthService) AuthCodeURL(state string) string {
	url_, err := url.Parse(c.AuthURL)
	if err != nil {
		panic("AuthURL malformed: " + err.Error())
	}
	q := url.Values{
		"response_type": {"code"},
		"client_id":     {c.ClientID},
		"redirect_uri":  {c.redirectURL()},
		"scope":         {c.Scope},
		"state":         {state},
	}.Encode()
	if url_.RawQuery == "" {
		url_.RawQuery = q
	} else {
		url_.RawQuery += "&" + q
	}
	return url_.String()
}

type Transport struct {
	*OAuthService
	*Token
	Transport http.RoundTripper
}

func (t *Transport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// ExchangeAuthorizationCode takes a code and gets access Token from the remote server.
// http://tools.ietf.org/html/draft-ietf-oauth-v2-23#section-4.2.1
func (t *Transport) ExchangeAuthorizationCode(code string) (*Token, error) {
	if t.OAuthService == nil {
		return nil, errors.New("no OAuthService supplied")
	}
	tok := new(Token)
	err := t.updateToken(tok, url.Values{
		"grant_type":   {"authorization_code"},
		"redirect_uri": {t.redirectURL()},
		"scope":        {t.Scope},
		"code":         {code},
	})
	if err != nil {
		return nil, err
	}
	return tok, nil
}

// ExchangeClientCredentials uses the client credentials and gets an access Token
// from the authorization server
// http://tools.ietf.org/html/draft-ietf-oauth-v2-30#section-4.4
func (t *Transport) ExchangeClientCredentials() (*Token, error) {
	if t.OAuthService == nil {
		return nil, errors.New("no OAuthService supplied")
	}
	tok := new(Token)
	err := t.updateToken(tok, url.Values{
		"grant_type": {"client_credentials"},
		"scope":      {t.Scope},
	})
	if err != nil {
		return nil, err
	}
	return tok, nil
}

// RoundTrip executes a single HTTP transaction using the Transport's
// Token as authorization headers.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.OAuthService == nil {
		return nil, errors.New("no OAuthService supplied")
	}
	if t.Token == nil {
		return nil, errors.New("no Token supplied")
	}

	// Refresh the Token if it has expired.
	if t.Expired() {
		if err := t.Refresh(); err != nil {
			return nil, err
		}
	}

	// Make the HTTP request.
	req.Header.Set("Authorization", "OAuth "+t.AccessToken)
	return t.transport().RoundTrip(req)
}

// Refresh renews the Transport's AccessToken using its RefreshToken.
func (t *Transport) Refresh() error {
	if t.OAuthService == nil {
		return errors.New("no OAuthService supplied")
	} else if t.Token == nil {
		return errors.New("no existing Token")
	}

	err := t.updateToken(t.Token, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {t.RefreshToken},
	})
	if err != nil {
		return err
	}
	return nil
}

func (t *Transport) updateToken(tok *Token, v url.Values) error {
	v.Set("client_id", t.ClientID)
	v.Set("client_secret", t.ClientSecret)
	tr := ExchangeTransport{AcceptHeader: t.AcceptHeader}
	r, err := (&http.Client{Transport: tr}).PostForm(t.TokenURL, v)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return errors.New("invalid response: " + r.Status)
	}
	var b struct {
		Access    string        `json:"access_token"`
		Refresh   string        `json:"refresh_token"`
		ExpiresIn time.Duration `json:"expires_in"`
	}
	buf := &bytes.Buffer{}
	io.Copy(buf, r.Body)
	contenttype := strings.Split(r.Header.Get("Content-Type"), "; ")

	switch contenttype[0] {
	case "application/json":
		if err = json.Unmarshal(buf.Bytes(), &b); err != nil {
			return err
		}
	case "application/xml":
		if err = xml.Unmarshal(buf.Bytes(), &b); err != nil {
			return err
		}

	case "text/plain":
		body := make([]byte, r.ContentLength)
		_, err = r.Body.Read(body)
		if err != nil {
			return err
		}
		vals, err := url.ParseQuery(string(body))
		if err != nil {
			return err
		}

		b.Access = vals.Get("access_token")
		b.Refresh = vals.Get("refresh_token")
		expires_in, err := strconv.ParseInt(vals.Get("expires"), 10, 64)
		if err != nil {
			return err
		}
		b.ExpiresIn = time.Duration(expires_in)

	default:
		return errors.New("Unknown token format")
	}
	tok.AccessToken = b.Access
	if len(b.Refresh) > 0 {
		tok.RefreshToken = b.Refresh
	}
	if b.ExpiresIn == 0 {
		tok.Expiry = time.Time{}
	} else {
		tok.Expiry = time.Now().Add(b.ExpiresIn * time.Second)
	}
	return nil
}

// AlternateTransport implements http.RoundTripper. It sets an additional header value
// for OAuth 2 providers that detect the Accept header.
type ExchangeTransport struct {
	AcceptHeader string
	Transport    http.RoundTripper
}

func (t ExchangeTransport) transport() http.RoundTripper {
	return http.DefaultTransport
}

func (t ExchangeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.AcceptHeader != "" {
		req.Header.Set("Accept", t.AcceptHeader)
	}
	return t.transport().RoundTrip(req)
}
