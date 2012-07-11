// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package main

import (
	"amp/model"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"sync"
)

// const url = "http://localhost:8080/?key=blah"
const url = "https://togethrat.appspot.com/?key=blah"

type Codec struct {
	m      sync.Mutex
	client *http.Client
	enc    *gob.Encoder
	encBuf *bytes.Buffer
	err    bool
	resp   io.ReadCloser
	dec    *gob.Decoder
}

func (c *Codec) WriteRequest(r *rpc.Request, body interface{}) (err os.Error) {
	defer c.m.Unlock()
	if err = c.enc.Encode(r); err != nil {
		c.err = true
		return
	}
	if err = c.enc.Encode(body); err != nil {
		c.err = true
		return
	}
	resp, err := c.client.Post(url, "raw", c.encBuf)
	if err != nil {
		c.err = true
		return
	}
	c.resp = resp.Body
	c.dec = gob.NewDecoder(resp.Body)
	return nil
}

func (c *Codec) ReadResponseHeader(r *rpc.Response) os.Error {
	c.m.Lock()
	defer c.m.Unlock()
	if c.err {
		return os.EOF
	}
	return c.dec.Decode(r)
}

func (c *Codec) ReadResponseBody(body interface{}) os.Error {
	if c.err {
		return os.EOF
	}
	return c.dec.Decode(body)
}

func (c *Codec) Close() os.Error {
	return c.resp.Close()
}

func Dial() *rpc.Client {
	httpClient := &http.Client{}
	buf := &bytes.Buffer{}
	codec := &Codec{
		client: httpClient,
		enc:    gob.NewEncoder(buf),
		encBuf: buf,
	}
	codec.m.Lock()
	return rpc.NewClientWithCodec(codec)
}

func main() {

	client := Dial()

	item := &model.Item{
		Parent1: "boo",
	}

	var key string

	err := client.Call("db.CreateItem", item, &key)
	if err != nil {
		log.Fatal("Error:", err)
	}

	fmt.Printf("Key: %v\n", key)

}
