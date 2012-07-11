// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package messaging

import (
	"amp/argo"
	"io"
)

const CurrentVersion = 1

// Packet Types
const (
	PingPacket = iota
	HelloPacket
	CookiePacket
	InitiatePacket
	MessagePacket
)

type Connection struct {

}

type Packet struct {
	typ uint64
}

func (conn *Connection) ReadPacket(stream io.Reader) (packet *Packet, err error) {
	decoder := argo.NewDecoder(stream)
	typ, err := decoder.ReadVarint()
	if err != nil {
		return
	}
	if typ > MessagePacket {
		return
	}
	packet = &Packet{
		typ: typ,
	}
	return
}
