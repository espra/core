// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package protocol

import (
	"amp/argo"
	"io"
	"os"
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

func (conn *Connection) ReadPacket(stream io.Reader) (packet *Packet, err os.Error) {
	decoder := &argo.Decoder{stream}
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
