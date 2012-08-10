// Public Domain (-) 2011-2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package db

import (
	"amp/zerodata"
	"sync"
)

var (
	clients       []*client
	defaultSecret []byte
	defaultURL    string
	mutex         sync.Mutex
)

type client struct {
	impl *zerodata.Client
}

func (c *client) Get() {

}

func (c *client) Release() {
	mutex.Lock()
	clients = append(clients, c)
	mutex.Unlock()
}

func Client() *client {
	mutex.Lock()
	defer mutex.Unlock()
	i := len(clients) - 1
	if i >= 0 {
		c := clients[i]
		clients = clients[:i]
		return c
	}
	return &client{zerodata.NewClient(defaultSecret, defaultURL)}
}

func NewClientWithOpts(secret []byte, url string) *client {
	return &client{zerodata.NewClient(secret, url)}
}

func SetDefaultOpts(secret []byte, url string) {
	defaultSecret = secret
	defaultURL = url
}
