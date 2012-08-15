// Public Domain (-) 2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package zerodata

import (
	"amp/dbi"
	"amp/tlsconf"
	"bytes"
	"encoding/gob"
	"errors"
	"io"
	"net/http"
	"sync"
)

type pad struct {
	buf  *bytes.Buffer
	enc  *gob.Encoder
	rbuf [512]byte
}

func newPad() *pad {
	buf := &bytes.Buffer{}
	return &pad{
		buf: buf,
		enc: gob.NewEncoder(buf),
	}
}

type Client struct {
	mutex  sync.Mutex
	pads   []*pad
	secret []byte
	url    string
	web    *http.Client
}

func (c *Client) Call(service string, req interface{}, resp interface{}) error {
	c.mutex.Lock()
	i := len(c.pads) - 1
	var p *pad
	if i >= 0 {
		p = c.pads[i]
		c.pads = c.pads[:i]
		c.mutex.Unlock()
	} else {
		c.mutex.Unlock()
		p = newPad()
	}
	defer func() {
		p.buf.Reset()
		c.mutex.Lock()
		c.pads = append(c.pads, p)
		c.mutex.Unlock()
	}()
	p.buf.Write(c.secret)
	err := p.enc.Encode(req)
	if err != nil {
		return err
	}
	r, err := c.web.Post(c.url+service, "rpc/gob", p.buf)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		n, err := r.Body.Read(p.rbuf[:])
		if err != nil && err != io.EOF {
			return err
		}
		return errors.New(string(p.rbuf[:n]))
	}
	_, err = r.Body.Read(p.rbuf[:1])
	if err != nil {
		return err
	}
	switch p.rbuf[0] {
	case '0':
		n, err := r.Body.Read(p.rbuf[:])
		if err != nil && err != io.EOF {
			return err
		}
		return errors.New(string(p.rbuf[:n]))
	case '1':
		dec := gob.NewDecoder(r.Body)
		dec.Decode(resp)
	case '2':
		return dbi.EntityNotFound
	case '3':
		return dbi.StopIteration
	case '4':
		return io.EOF
	default:
		return errors.New("rpc: invalid gob response header")
	}
	return nil
}

func NewClient(serverURL string, secretKey []byte, transport http.RoundTripper) *Client {
	if transport == nil {
		transport = &http.Transport{TLSClientConfig: tlsconf.Config}
	}
	return &Client{
		secret: secretKey,
		url:    serverURL,
		web:    &http.Client{Transport: transport},
	}
}
