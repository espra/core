// Public Domain (-) 2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package zerodata

import (
	"appengine"
	"appengine/datastore"
	"crypto/subtle"
	"dbi"
	"fmt"
	"net"
	"net/http"
)

func getKey(ctx appengine.Context, key *dbi.Key) *datastore.Key {
	var parent *datastore.Key
	if key.Parent != nil {
		parent = getKey(ctx, key.Parent)
	}
	return datastore.NewKey(ctx, key.Kind, key.StringID, key.IntID, parent)
}

func retKey(key *datastore.Key) *dbi.Key {
	if key == nil {
		return nil
	}
	return &dbi.Key{
		Kind:     key.Kind(),
		StringID: key.StringID(),
		IntID:    key.IntID(),
		Parent:   retKey(key.Parent()),
	}
}

func AllocateIDs(ctx appengine.Context, req *dbi.AllocRequest, alloc *dbi.Allocation) error {
	low, high, err := datastore.AllocateIDs(ctx, req.Kind, getKey(ctx, req.Parent), req.Amount)
	if err != nil {
		return err
	}
	alloc.Low = low
	alloc.High = high
	return nil
}

func Delete(ctx appengine.Context, key *dbi.Key, ok bool) error {
	return datastore.Delete(ctx, getKey(ctx, key))
}

func DeleteMulti(ctx appengine.Context, keys []*dbi.Key, ok bool) error {
	xkeys := make([]*datastore.Key, len(keys))
	for idx, key := range keys {
		xkeys[idx] = getKey(ctx, key)
	}
	return datastore.DeleteMulti(ctx, xkeys)
}

func Get(ctx appengine.Context, key *dbi.Key, entity *datastore.PropertyList) error {
	return datastore.Get(ctx, getKey(ctx, key), entity)
}

func GetMulti(ctx appengine.Context, keys []*dbi.Key, entities *dbi.EntityList) error {
	xkeys := make([]*datastore.Key, len(keys))
	for idx, key := range keys {
		xkeys[idx] = getKey(ctx, key)
	}
	return datastore.GetMulti(ctx, xkeys, entities.List)
}

type KeyValue struct {
	Key   *dbi.Key
	Value datastore.PropertyList
}

func Put(ctx appengine.Context, kv KeyValue, rkey *dbi.Key) error {
	key, err := datastore.Put(ctx, getKey(ctx, kv.Key), &kv.Value)
	if err != nil {
		return err
	}
	rkey = retKey(key)
	return nil
}

func PutMulti(ctx appengine.Context, kvlist []*dbi.KeyValue, rkeylist *dbi.KeyList) error {
	keys := make([]*datastore.Key, len(kvlist))
	entities := make([]dbi.Entity, len(kvlist))
	for idx, kv := range kvlist {
		keys[idx] = getKey(ctx, kv.Key)
		entities[idx] = kv.Value
	}
	keys, err := datastore.PutMulti(ctx, keys, entities)
	if err != nil {
		return err
	}
	rkeys := make([]*dbi.Key, len(keys))
	for idx, key := range keys {
		rkeys[idx] = retKey(key)
	}
	rkeylist.List = rkeys
	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	// Provide some basic info for GET requests.
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, "Zerodata API Endpoint.")
		return
	}
	// Instantiate the App Engine Context object.
	ctx := appengine.NewContext(r)
	// Do some sanity checks when in production.
	if !appengine.IsDevAppServer() {
		// Ensure that we are only being accessed over HTTPS.
		if r.TLS == nil {
			raise(400, ctx, w, "Endpoint must only be accessed over HTTPS in production.")
			return
		}
		// Check the IP address against the whitelist.
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			raise(400, ctx, w, "Couldn't parse the remote address: %s", err)
			return
		}
		if !isWhitelisted(ctx, host) {
			raise(400, ctx, w, "Sorry, the IP address %q hasn't been whitelisted.", host)
			return
		}
	}
	// Check the secret key.
	key := make([]byte, secretLength)
	n, err := r.Body.Read(key)
	if err != nil || n != secretLength {
		raise(401, ctx, w, "Sorry, invalid secret key.")
		return
	}
	if subtle.ConstantTimeCompare(key, secret) != 1 {
		raise(401, ctx, w, "Sorry, invalid secret key.")
		return
	}
	if len(r.URL.Path) == 0 {
		http.NotFound(w, r)
		return
	}
	// Delegate the request to the RPC handler.
	handleRPC(ctx, w, r)
}

func raise(status int, ctx appengine.Context, w http.ResponseWriter, s string, args ...interface{}) {
	ctx.Errorf(s, args...)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprint(w, "ERROR: ")
	fmt.Fprintf(w, s, args...)
}

func init() {
	// Validate the secret key settings.
	if secretLength == 0 {
		panic("the secret key cannot be empty.")
	}
	if secretLength != len(secret) {
		panic("the secret key length does not match the `secretLength` variable.")
	}
	// Register RPC services.
	Register("allocate-ids", AllocateIDs)
	Register("delete", Delete)
	Register("delete-multi", DeleteMulti)
	Register("get", Get)
	Register("get-multi", GetMulti)
	Register("put", Put)
	Register("put-multi", PutMulti)
	// Register the root HTTP handler.
	http.HandleFunc("/", handler)
}
