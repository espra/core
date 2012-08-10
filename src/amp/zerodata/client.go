// Public Domain (-) 2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package zerodata

import (
	// "amp/dbi"
	"bytes"
	"encoding/gob"
	"errors"
	"io"
	"net/http"
	"sync"
)

type ZeroDataClient struct {
	buf    *bytes.Buffer
	enc    *gob.Encoder
	mutex  sync.Mutex
	rbuf   [512]byte
	secret []byte
	url    string
	web    *http.Client
}

func (c *ZeroDataClient) Call(service string, req interface{}, resp interface{}) error {
	c.mutex.Lock()
	c.buf.Reset()
	c.buf.Write(c.secret)
	err := c.enc.Encode(req)
	if err != nil {
		c.mutex.Unlock()
		return err
	}
	r, err := c.web.Post(c.url+service, "rpc/gob", c.buf)
	c.mutex.Unlock()
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		n, err := r.Body.Read(c.rbuf[:])
		if err != nil && err != io.EOF {
			return err
		}
		return errors.New(string(c.rbuf[:n]))
	}
	_, err = r.Body.Read(c.rbuf[:1])
	if err != nil {
		return err
	}
	switch c.rbuf[0] {
	case '0':
		n, err := r.Body.Read(c.rbuf[:])
		if err != nil && err != io.EOF {
			return err
		}
		return errors.New(string(c.rbuf[:n]))
	case '1':
		dec := gob.NewDecoder(r.Body)
		dec.Decode(resp)
	default:
		return errors.New("rpc: invalid gob response header")
	}
	return nil
}

func NewClient(secret []byte, url string) *ZeroDataClient {
	buf := &bytes.Buffer{}
	return &ZeroDataClient{
		buf:    buf,
		enc:    gob.NewEncoder(buf),
		secret: secret,
		url:    url,
		web:    &http.Client{},
	}
}
