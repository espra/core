// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package livequery

import (
	"amp/refmap"
	"strings"
	"sync"
	"time"
)

type Notification struct {
	item string
	qid  string
}

type Query struct {
	keys []string
	seen int64
}

type Session struct {
	listener chan *Notification
	queries  map[string]*Query
}

type Subscription struct {
	sqidRef uint64
	tally   int
}

type PubSub struct {
	mutex    sync.RWMutex
	refmap   *refmap.Map
	sessions map[string]*Session
	subs     map[string][]*Subscription
}

func (pubsub *PubSub) Listen(sid string) map[string][]string {
	results := make(map[string][]string)
	return results
}

func (pubsub *PubSub) Publish(item string, keys []string) {
	tally := len(keys)
	if tally == 0 {
		return
	}
	data := pubsub.subs
	counts := make(map[*Subscription]int)
	pubsub.mutex.RLock()
	for _, key := range keys {
		if subs, found := data[key]; found {
			for _, sub := range subs {
				counts[sub] += 1
			}
		}
	}
	pubsub.mutex.RUnlock()
	for sub, count := range counts {
		if sub.tally == count {
			go func() {
			}()
		}
	}
}

func (pubsub *PubSub) Subscribe(sqid string, keys []string, keys2 []string) {

	splitSqid := strings.Split(sqid, ":", 2)
	if len(splitSqid) != 2 {
		return
	}

	sid := splitSqid[0]
	qid := splitSqid[1]

	tally := len(keys)
	keys2Len := len(keys2)
	refCount := tally + keys2Len

	if refCount == 0 {
		return
	}
	if keys2Len > 0 {
		tally += 1
	}

	sqidRef := pubsub.refmap.Incref(sqid, refCount)

	for _, key := range keys2 {
		keys = append(keys, key)
	}

	now := time.Seconds()
	data := pubsub.subs
	sessions := pubsub.sessions
	pubsub.mutex.Lock()
	session, found := sessions[sid]

	if found {
		query, found := session.queries[qid]
		if found {
			query.seen = now
		} else {
			query := &Query{keys: keys, seen: now}
			session.queries[qid] = query
		}
	} else {
		listener := make(chan *Notification, 100)
		queries := make(map[string]*Query)
		query := &Query{keys: keys, seen: now}
		queries[qid] = query
		session := &Session{
			listener: listener,
			queries:  queries,
		}
		sessions[sid] = session
	}

	for _, key := range keys {
		sub := &Subscription{
			sqidRef: sqidRef,
			tally:   tally,
		}
		subs, found := data[key]
		if found {
			data[key] = append(subs, sub)
		} else {
			data[key] = []*Subscription{sub}
		}
	}

	pubsub.mutex.Unlock()

}

func New() *PubSub {
	refmap := refmap.New()
	sessions := make(map[string]*Session)
	subs := make(map[string][]*Subscription)
	return &PubSub{
		refmap:   refmap,
		sessions: sessions,
		subs:     subs,
	}
}
