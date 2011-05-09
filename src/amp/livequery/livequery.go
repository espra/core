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
	keys    []string
	qid     string
	seen    int64
	session *Session
}

type Session struct {
	listener chan *Notification
	mutex    sync.Mutex
	seen     int64
}

type Subscription struct {
	sqidRef uint64
	tally   int
}

type PubSub struct {
	mutex         sync.RWMutex
	queries       map[uint64]*Query
	refmap        *refmap.Map
	sessions      map[string]*Session
	subscriptions map[string][]*Subscription
}

// -----------------------------------------------------------------------------
// Listen
// -----------------------------------------------------------------------------

func (pubsub *PubSub) Listen(sid string, qids []string, timeout int64) (result map[string][]string, refresh map[string]int, ok bool) {

	sqids := make([]string, len(qids))
	for idx, qid := range qids {
		sqids[idx] = sid + ":" + qid
	}

	refresh = make(map[string]int)
	refs := pubsub.refmap.MultiGet(sqids...)

	for idx, ref := range refs {
		if ref == refmap.Zero {
			refresh[qids[idx]] = 1
		}
	}

	if len(refresh) > 0 {
		return nil, refresh, false
	}

	now := time.Seconds()
	pubsub.mutex.Lock()

	session, found := pubsub.sessions[sid]
	if !found {
		pubsub.mutex.Unlock()
		return
	}

	listener := session.listener
	queries := pubsub.queries

	for idx, ref := range refs {
		if query, found := queries[ref]; found {
			query.seen = now
		} else {
			refresh[qids[idx]] = 1
		}
	}

	pubsub.mutex.Unlock()
	if len(refresh) > 0 {
		return nil, refresh, false
	}

	waiting := true
	session.mutex.Lock()
	defer session.mutex.Unlock()

	go func() {
		<-time.After(timeout)
		if waiting {
			notification := &Notification{}
			listener <- notification
		}
	}()

	notification := <-listener
	waiting = false
	result = make(map[string][]string)

	if notification.item == "" {
		return result, nil, true
	}

	result[notification.qid] = []string{notification.item}

	var qid string

	for i := 0; i < len(listener); i++ {
		notification = <-listener
		qid = notification.qid
		resp, found := result[qid]
		if found {
			result[qid] = append(resp, notification.item)
		} else {
			result[qid] = []string{notification.item}
		}
	}

	return result, nil, true

}

// -----------------------------------------------------------------------------
// Publish
// -----------------------------------------------------------------------------

func (pubsub *PubSub) Publish(item string, keys []string) {

	tally := len(keys)
	if tally == 0 {
		return
	}

	queries := pubsub.queries
	subscriptions := pubsub.subscriptions
	counts := make(map[*Subscription]int)
	pubsub.mutex.RLock()

	for _, key := range keys {
		if subs, found := subscriptions[key]; found {
			for _, sub := range subs {
				counts[sub] += 1
			}
		}
	}

	resp := make(map[*Query]int)
	for sub, count := range counts {
		if sub.tally == count {
			resp[queries[sub.sqidRef]] = 1
		}
	}

	pubsub.mutex.RUnlock()

	for query, _ := range resp {
		go func(query *Query) {
			notification := &Notification{
				item: item,
				qid:  query.qid,
			}
			query.session.listener <- notification
		}(query)
	}

}

// -----------------------------------------------------------------------------
// Subscribe
// -----------------------------------------------------------------------------

func (pubsub *PubSub) Subscribe(sqid string, keys []string, keys2 []string) {

	splitSqid := strings.Split(sqid, ":", 2)
	if len(splitSqid) != 2 {
		return
	}

	sid := splitSqid[0]
	qid := splitSqid[1]

	tally := len(keys)
	keys2Len := len(keys2)

	if tally+keys2Len == 0 {
		return
	}

	if keys2Len > 0 {
		tally += 1
		keys = append(keys, keys2...)
	}

	now := time.Seconds()
	sqidRef := pubsub.refmap.Create(sqid)
	queries := pubsub.queries
	sessions := pubsub.sessions
	subscriptions := pubsub.subscriptions

	pubsub.mutex.Lock()
	session, found := sessions[sid]

	// Create a new session if one doesn't already exist for the Session ID.
	if !found {
		listener := make(chan *Notification, 100)
		session = &Session{
			listener: listener,
			seen:     now,
		}
		sessions[sid] = session
	}

	queries[sqidRef] = &Query{
		keys:    keys,
		qid:     qid,
		seen:    now,
		session: session,
	}

	sub := &Subscription{
		sqidRef: sqidRef,
		tally:   tally,
	}

	for _, key := range keys {
		subs, found := subscriptions[key]
		if found {
			subscriptions[key] = append(subs, sub)
		} else {
			subscriptions[key] = []*Subscription{sub}
		}
	}

	pubsub.mutex.Unlock()

}

// -----------------------------------------------------------------------------
// Cleanup
// -----------------------------------------------------------------------------

func (pubsub *PubSub) Cleanup(expire int64, interval int64) {

	blankSubs := make([]*Subscription, 0)
	queries := pubsub.queries
	refmap := pubsub.refmap
	sessions := pubsub.sessions
	subscriptions := pubsub.subscriptions

	for {
		now := time.Seconds()
		pubsub.mutex.Lock()

		removeKeys := make(map[string]int)
		removeQueries := make(map[uint64]*Query)
		removeSessions := make(map[string]*Session)

		// Loop through the queries and find the ones we haven't seen in a
		// while. Also, gather together the keys which hold subscriptions for
		// those queries.
		for ref, query := range queries {
			if now-query.seen < expire {
				continue
			}
			removeQueries[ref] = query
			for _, key := range query.keys {
				removeKeys[key] += 1
			}
		}

		// Remove those queries.
		for ref, query := range removeQueries {
			queries[ref] = query, false
		}

		// Loop through the affected subscription keys and modify the
		// subscriptions appropriately.
		for key, remsize := range removeKeys {
			subs := subscriptions[key]
			subsize := len(subs)
			if remsize == subsize {
				subscriptions[key] = blankSubs, false
			} else {
				newsubs := make([]*Subscription, subsize-remsize)
				idx := 0
				for _, sub := range subs {
					if _, found := removeQueries[sub.sqidRef]; !found {
						newsubs[idx] = sub
						idx += 1
					}
				}
			}
		}

		// Look through the sessions and find the ones we haven't seen in a
		// while.
		for sid, session := range sessions {
			if now-session.seen > expire {
				removeSessions[sid] = session
			}
		}

		// Remove those sessions.
		for sid, session := range removeSessions {
			sessions[sid] = session, false
		}

		pubsub.mutex.Unlock()

		// Delete the refmap references for the various sqids.
		for ref, _ := range removeQueries {
			refmap.DeleteRef(ref)
		}

		<-time.After(interval)
	}

}

// -----------------------------------------------------------------------------
// Initialiser
// -----------------------------------------------------------------------------

func New() *PubSub {
	refmap := refmap.New()
	queries := make(map[uint64]*Query)
	sessions := make(map[string]*Session)
	subcriptions := make(map[string][]*Subscription)
	return &PubSub{
		queries:       queries,
		refmap:        refmap,
		sessions:      sessions,
		subscriptions: subcriptions,
	}
}
