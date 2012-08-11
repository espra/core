// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package amp

import (
	"amp/rpc"
	"io"
)

type Context struct {
	// Error
	Header         rpc.Header
	ResponseHeader rpc.Header
	Streams        []io.ReadCloser
}

func (ctx *Context) Call(method string, args ...interface{}) (req *Request) {
	return
}

func (ctx *Context) Dial(identity string) (req *Request) {
	return
}

func (ctx *Context) DialHTTPS(url string) (req *Request) {
	return
}

func (ctx *Context) Error() {
}

func (ctx *Context) SetStream(stream io.ReadCloser) {
	if ctx.Streams == nil {
		ctx.Streams = []io.ReadCloser{stream}
	} else {
		ctx.Streams = append(ctx.Streams, stream)
	}
}

type Request struct {
	Header rpc.Header
}

func (req *Request) Return(vals ...interface{}) (err error) {
	return nil
}

func Call() {
}

func Register() {
}

// func Hello(ctx *web.Context, files *web.Files) {
// }

// web.GetCookie(ctx)

// func (db *DB) CreateItem(ctx, user string, item *Item) {
// 	if !db.CheckUser(ctx, user) {
// 		return
// 	}
// 	key = NewKey(parent=user)
// 	db.Put(key, item)
// }

// func (db *DB) CheckUser(ctx, user) (ok bool) {
// 	if ctx.Header.GetString("__user") != user {
// 		ctx.Error("AuthError")
// 		return false
// 	}
// }

// in nodule memcache:

// func (api *API) Get(ctx, key string) {
// 	key = Namespace(ctx, key)
// 	return ctx.Call(memcacheRef, "memcache.Get", key)
// }

// in nodule memcacheRef:

// func (cache *MemCache) Get(ctx, key string) {
// 	return cache.values[key]
// }

// Register(&Memcache{})
// Register(&Facebook{Secret: "SECRET", AppID: "Blah"})
// RegisterName("fb", &Facebook{Secret: "SECRET", AppID: "Blah"})
