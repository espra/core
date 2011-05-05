// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package livequery

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

const timeout int64 = 1000000000

func listen(t *testing.T, pubsub *PubSub, sid string, qids ...string) {
	updates, refresh, ok := pubsub.Listen(sid, qids, timeout)
	if !ok {
		t.Errorf("Failed to get updates for %s:%v", sid, qids)
		if len(refresh) != 0 {
			t.Errorf("Refresh for %s:%v -- %v", sid, qids, refresh)
		}
		return
	}

	fmt.Printf("Responses for %s:%v = %v\n", sid, qids, updates)
}

func TestLiveQuery(t *testing.T) {

	pubsub := New()
	runtime.GOMAXPROCS(4)

	// Make various subscriptions.
	pubsub.Subscribe("tav:1", []string{"to:#espians"}, []string{})

	pubsub.Subscribe("tav:7", []string{"to:#bigsociety"},
		[]string{"from:indy", "from:evangineer"})

	pubsub.Subscribe("b0ggle:12", []string{"to:#espians", "ampify"},
		[]string{})

	pubsub.Subscribe("t:5", []string{"android"},
		[]string{"from:tav", "from:evangineer"})

	// Publish new items.
	pubsub.Publish("100",
		[]string{"from:tav", "to:#espians", "android", "sucks"})

	pubsub.Publish("101",
		[]string{"from:tav", "to:#espians", "ampify", "compiles"})

	pubsub.Publish("102",
		[]string{"from:indy", "to:#bigsociety", "to:#espians", "see", "togethr"})

	pubsub.Publish("103",
		[]string{"from:indy", "to:#espians", "when", "shall", "we", "launch"})

	pubsub.Publish("104",
		[]string{"from:evangineer", "to:#android", "got", "the", "device"})

	pubsub.Publish("105",
		[]string{"from:evangineer", "to:#bigsociety", "cameron", "failed"})

	go pubsub.Cleanup(timeout/1000000000, timeout/10)

	// Listen for them.
	go listen(t, pubsub, "tav", "1", "7")
	go listen(t, pubsub, "b0ggle", "12")
	go listen(t, pubsub, "t", "5")
	go listen(t, pubsub, "tav", "1", "7")

	<-time.After(timeout * 3)

	fmt.Printf("Queries: %#v\n", pubsub.queries)
	fmt.Printf("Sessions: %#v\n", pubsub.sessions)
	fmt.Printf("Subs: %#v\n", pubsub.subscriptions)

}
