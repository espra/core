// Public Domain (-) 2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package zerodata

import (
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"sync"
	"time"
)

const defaultExpiry = 30 * time.Second

var ipCache = map[string]StateInfo{}
var ipLock sync.Mutex

type StateInfo struct {
	state   bool
	expires time.Time
}

type WhitelistIP struct {
	T time.Time
}

func cacheIP(addr string, state bool, now time.Time) {
	ipLock.Lock()
	ipCache[addr] = StateInfo{state: state, expires: now.Add(defaultExpiry)}
	ipLock.Unlock()
}

// isWhitelisted returns whether the given IP address has been whitelisted.
func isWhitelisted(ctx appengine.Context, addr string) bool {
	if addr == "" {
		return false
	}
	ipLock.Lock()
	info, exists := ipCache[addr]
	ipLock.Unlock()
	now := time.Now()
	if exists && info.expires.After(now) {
		return info.state
	}
	memkey := "w:" + addr
	item, err := memcache.Get(ctx, memkey)
	if err == nil {
		if len(item.Value) == 1 {
			switch item.Value[0] {
			case '0':
				cacheIP(addr, false, now)
				return false
			case '1':
				cacheIP(addr, true, now)
				return true
			}
		}
	}
	w := WhitelistIP{}
	key := datastore.NewKey(ctx, "W", addr, 0, nil)
	err = datastore.Get(ctx, key, &w)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			memcache.Set(ctx, &memcache.Item{Key: memkey, Value: []byte{'0'}})
			cacheIP(addr, false, now)
			return false
		}
		ctx.Criticalf("Unable to look up whitelisted address %q from the datastore: %s", addr, err)
		return false
	}
	memcache.Set(ctx, &memcache.Item{Key: memkey, Value: []byte{'1'}})
	cacheIP(addr, true, now)
	return true
}
