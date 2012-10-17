// Public Domain (-) 2011-2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package structure

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"github.com/dchest/siphash"
	"sort"
)

const dictSize = 4096

// Dict is a minimalist hash table for []byte keys. It uses SipHash to provide
// built-in defense against adversarial keys.
type Dict struct {
	data [dictSize]*dictNode
	k0   uint64
	k1   uint64
	size int64
}

type dictNode struct {
	k    []byte
	v    interface{}
	next *dictNode
}

func (d *Dict) Get(key []byte) (interface{}, bool) {
	node := d.data[siphash.Hash(d.k0, d.k1, key)%dictSize]
	if node == nil {
		return nil, false
	}
	for {
		if !bytes.Equal(node.k, key) {
			node = node.next
			if node == nil {
				return nil, false
			}
			continue
		}
		return node.v, true
	}
	panic("unreachable code")
}

func (d *Dict) Delete(key []byte) {
	p := siphash.Hash(d.k0, d.k1, key) % dictSize
	node := d.data[p]
	if node == nil {
		return
	}
	if bytes.Equal(node.k, key) {
		d.data[p] = nil
		d.size -= 1
		return
	}
	next := node.next
	for {
		if next == nil {
			return
		}
		if bytes.Equal(next.k, key) {
			node.next = next.next
			d.size -= 1
			return
		}
		next = node.next
	}
	panic("unreachable code")
}

func (d *Dict) FillCount() (count int) {
	for i := 0; i < dictSize; i++ {
		if d.data[i] != nil {
			count += 1
		}
	}
	return
}

func (d *Dict) Keys() (keys [][]byte) {
	var node *dictNode
	for i := 0; i < dictSize; i++ {
		node = d.data[i]
		for node != nil {
			keys = append(keys, node.k)
			node = node.next
		}
	}
	return
}

func (d *Dict) Set(key []byte, value interface{}) {
	p := siphash.Hash(d.k0, d.k1, key) % dictSize
	node := d.data[p]
	if node == nil {
		d.data[p] = &dictNode{
			k:    key,
			v:    value,
			next: nil,
		}
		d.size += 1
		return
	}
	for {
		if bytes.Equal(node.k, key) {
			node.v = value
			return
		}
		if node.next != nil {
			node = node.next
			continue
		}
		node.next = &dictNode{
			k:    key,
			v:    value,
			next: nil,
		}
		d.size += 1
		return
	}
}

func (d *Dict) Size() int64 {
	return d.size
}

func NewDict() (*Dict, error) {
	k := make([]byte, 16)
	_, err := rand.Read(k)
	if err != nil {
		return nil, err
	}
	k0 := binary.BigEndian.Uint64(k[:8])
	k1 := binary.BigEndian.Uint64(k[8:16])
	return &Dict{
		k0: k0,
		k1: k1,
	}, nil
}

func SortedKeys(dict map[string]string) (keys []string) {
	keys = make([]string, len(dict))
	i := 0
	for key, _ := range dict {
		keys[i] = key
		i += 1
	}
	sort.StringSlice(keys).Sort()
	return
}
