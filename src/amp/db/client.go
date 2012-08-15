// Public Domain (-) 2011-2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package db

import (
	"amp"
	"amp/zerodata"
	"sync"
)

var (
	clients       []*Client
	defaultSecret []byte
	defaultURL    string
	mutex         sync.Mutex
)

type Client struct {
	ctx  *amp.Context
	impl *zerodata.Client
	rel  bool
}

func (c *Client) Get() {

}

func (c *Client) Release() {
	if !c.rel {
		return
	}
	mutex.Lock()
	clients = append(clients, c)
	mutex.Unlock()
}

func NewClient(ctx *amp.Context) *Client {
	mutex.Lock()
	defer mutex.Unlock()
	i := len(clients) - 1
	if i >= 0 {
		c := clients[i]
		clients = clients[:i]
		c.ctx = ctx
		return c
	}
	return &Client{
		ctx:  ctx,
		impl: zerodata.NewClient(defaultSecret, defaultURL),
		rel:  true,
	}
}

func NewClientWithOpts(ctx *amp.Context, secret []byte, url string) *Client {
	return &Client{
		ctx:  ctx,
		impl: zerodata.NewClient(secret, url),
		rel:  false,
	}
}

func SetDefaultOpts(secret []byte, url string) {
	defaultSecret = secret
	defaultURL = url
}
