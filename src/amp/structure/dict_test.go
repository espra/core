// Public Domain (-) 2011-2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package structure

import (
	"testing"
)

type kv struct {
	k []byte
	v string
}

var benchKeys = []kv{
	kv{[]byte("trust-maps"), "service-1"},
	kv{[]byte("login"), "service-2"},
	kv{[]byte("pecu.allocation"), "service-3"},
	kv{[]byte("pecu.payout"), "service-4"},
	kv{[]byte("appreciate"), "service-5"},
	kv{[]byte("memcache"), "service-6"},
	kv{[]byte("taskqueue"), "service-7"},
	kv{[]byte("reverse-geocoder"), "service-8"},
	kv{[]byte("image.thumbnail"), "service-9"},
	kv{[]byte("language.detect"), "service-10"},
}

type kvs struct {
	k string
	v string
}

var benchStringKeys = []kvs{
	kvs{"trust-maps", "service-1"},
	kvs{"login", "service-2"},
	kvs{"pecu.allocation", "service-3"},
	kvs{"pecu.payout", "service-4"},
	kvs{"appreciate", "service-5"},
	kvs{"memcache", "service-6"},
	kvs{"taskqueue", "service-7"},
	kvs{"reverse-geocoder", "service-8"},
	kvs{"image.thumbnail", "service-9"},
	kvs{"language.detect", "service-10"},
}

func BenchmarkDictGet(b *testing.B) {
	b.StopTimer()
	dict, _ := NewDict()
	var i byte
	for i = 0; i < 255; i++ {
		for _, elem := range benchKeys {
			dict.Set(append(elem.k, i), elem.v)
		}
	}
	var x interface{}
	var ok bool
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _, elem := range benchKeys {
			x, ok = dict.Get(elem.k)
		}
	}
	_, _ = x, ok
}

func BenchmarkMapGet(b *testing.B) {
	b.StopTimer()
	dict := map[string]interface{}{}
	var i byte
	for i = 0; i < 255; i++ {
		for _, elem := range benchKeys {
			dict[string(append(elem.k, i))] = elem.v
		}
	}
	var x interface{}
	var ok bool
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _, elem := range benchKeys {
			x, ok = dict[string(elem.k)]
		}
	}
	_, _ = x, ok
}

func BenchmarkDictStringGet(b *testing.B) {
	b.StopTimer()
	dict, _ := NewDict()
	var i byte
	for i = 0; i < 255; i++ {
		for _, elem := range benchStringKeys {
			dict.Set(append([]byte(elem.k), i), elem.v)
		}
	}
	var x interface{}
	var ok bool
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _, elem := range benchStringKeys {
			x, ok = dict.Get([]byte(elem.k))
		}
	}
	_, _ = x, ok
}

func BenchmarkMapStringGet(b *testing.B) {
	b.StopTimer()
	dict := map[string]interface{}{}
	var i byte
	for i = 0; i < 255; i++ {
		for _, elem := range benchKeys {
			dict[string(append(elem.k, i))] = elem.v
		}
	}
	var x interface{}
	var ok bool
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _, elem := range benchStringKeys {
			x, ok = dict[elem.k]
		}
	}
	_, _ = x, ok
}

func BenchmarkDictSet(b *testing.B) {
	b.StopTimer()
	dict, _ := NewDict()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _, elem := range benchKeys {
			dict.Set(elem.k, elem.v)
		}
	}
}

func BenchmarkMapSet(b *testing.B) {
	b.StopTimer()
	dict := map[string]interface{}{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _, elem := range benchKeys {
			dict[string(elem.k)] = elem.v
		}
	}
}

func BenchmarkDictStringSet(b *testing.B) {
	b.StopTimer()
	dict, _ := NewDict()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _, elem := range benchStringKeys {
			dict.Set([]byte(elem.k), elem.v)
		}
	}
}

func BenchmarkMapStringSet(b *testing.B) {
	b.StopTimer()
	dict := map[string]interface{}{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _, elem := range benchStringKeys {
			dict[elem.k] = elem.v
		}
	}
}

func BenchmarkStringCopy(b *testing.B) {
	b.StopTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _, elem := range benchStringKeys {
			x := []byte(elem.k)
			_ = x
		}
	}
}

func TestSortedKeys(t *testing.T) {
	dict := map[string]string{
		"tav":      "espian",
		"james":    "arthur",
		"sofia":    "bustamante",
		"mamading": "ceesay",
	}
	keys := SortedKeys(dict)
	expected := []string{"james", "mamading", "sofia", "tav"}
	for idx, key := range keys {
		if key != expected[idx] {
			t.Errorf("Unexpected item at position %d: %s (expected: %s)",
				idx, key, expected[idx])
		}
	}

}
