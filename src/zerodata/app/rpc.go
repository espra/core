// Public Domain (-) 2011-2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package zerodata

import (
	"appengine"
	"appengine/datastore"
	"encoding/gob"
	"io"
	"net/http"
	"reflect"
)

var ctxType = reflect.TypeOf((*appengine.Context)(nil)).Elem()
var errorType = reflect.TypeOf((*error)(nil)).Elem()

func handleRPC(ctx appengine.Context, w http.ResponseWriter, r *http.Request) {

	name := r.URL.Path[1:]
	s, exists := services[name]

	if !exists {
		http.NotFound(w, r)
		return
	}

	defer func() {
		if e := recover(); e != nil {
			raise(500, ctx, w, "internal server error: %s", e)
		}
	}()

	var iv reflect.Value
	if s.inPtr {
		iv = reflect.New(s.in.Elem())
	} else {
		iv = reflect.New(s.in)
	}

	dec := gob.NewDecoder(r.Body)
	err := dec.Decode(iv.Interface())

	if err != nil {
		panic("bad request: couldn't decode the input parameter: " + err.Error())
	}

	if !s.inPtr {
		iv = iv.Elem()
	}

	ov := reflect.New(s.out.Elem())
	ev := s.function.Call([]reflect.Value{reflect.ValueOf(ctx), iv, ov})

	w.Header().Set("Content-Type", "rpc/gob")

	if ev[0].IsNil() {
		w.Write([]byte{'1'})
		enc := gob.NewEncoder(w)
		enc.Encode(ov.Interface())
	} else {
		switch ev[0].Interface().(error) {
		case datastore.ErrNoSuchEntity:
			w.Write([]byte{'2'})
		case datastore.Done:
			w.Write([]byte{'3'})
		case io.EOF:
			w.Write([]byte{'4'})
		default:
			w.Write([]byte{'0'})
		}
		w.Write([]byte(err.Error()))
	}

}

type service struct {
	function reflect.Value
	in       reflect.Type
	inPtr    bool
	out      reflect.Type
}

var services = map[string]*service{}

func register(name string, v interface{}) *service {
	rv := reflect.ValueOf(v)
	rt := rv.Type()
	if rt.Kind() != reflect.Func {
		panic("rpc: attempted to register `" + rt.Kind().String() + "` object as `" + name + "`")
	}
	if rt.NumIn() != 3 {
		panic("rpc: the function for the " + name + " service needs to take three arguments")
	}
	if rt.In(0) != ctxType {
		panic("rpc: the first argument for `" + name + "` needs to be *rpc.Context")
	}
	if rt.NumOut() != 1 || rt.Out(0) != errorType {
		panic("rpc: the return argument for `" + name + "` needs to be error")
	}
	s := &service{
		function: rv,
		in:       rt.In(1),
		inPtr:    rt.In(1).Kind() == reflect.Ptr,
		out:      rt.In(2),
	}
	services[name] = s
	return s
}

func Register(name string, v interface{}) *service {
	return register(name, v)
}

type Namespace string

func (ns Namespace) Register(name string, v interface{}) *service {
	return register(string(ns)+"."+name, v)
}
