// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// The redis package provides a client library to interact with redis servers.
package redis

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	tcpConnection  = 0
	unixConnection = 1
	GET            = "GET"
	MGET           = "MGET"
	PING           = "PING"
	SET            = "SET"
	TTL            = "TTL"
)

var (
	MaxConnections  int
	OpenConnections int
	connections     = make(map[string]client, 100)
)

type client struct {
	address        string
	connectionType int
	multi          bool
	connected      bool
}

type keyspace struct {
	connected bool
}

func (keyspace *keyspace) Get(namespace string) *client {
	address := ""
	return &client{
		address:        address,
		connectionType: tcpConnection,
		multi:          false,
	}
}

func Keyspace(servers string) *keyspace {
	return &keyspace{}
}

// The ``Client`` constructor takes a single optional string parameter of the
// address of the redis server to connect to.
//
// If the address parameter is left out, it defaults to a TCP connection to
// ``localhost:6379``.
func Client(addr ...string) *client {

	var address string
	var connectionType int

	addrSlice := []string(addr)
	if len(addrSlice) > 0 {
		address = addrSlice[0]
		if strings.HasPrefix(address, "unix:") {
			address = address[5:]
			connectionType = unixConnection
		}
	} else {
		address = "localhost:6379"
		connectionType = tcpConnection
	}

	return &client{
		address:        address,
		connectionType: connectionType,
		multi:          false,
	}

}

// func (client *client) Connect() {
// 	var la, ra *net.TCPAddr
// 	if laddr != "" {
// 		if la, err = net.ResolveTCPAddr(laddr); err != nil {
// 			goto Error
// 		}
// 	}
// 	if raddr != "" {
// 		if ra, err = net.ResolveTCPAddr(raddr); err != nil {
// 			goto Error
// 		}
// 	}
// 	c, err := net.DialTCP(net, la, ra)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return c, nil
// }

var BlankArgs = [][]byte{}

func (client *client) Send(command string, args [][]byte) (response []byte, err os.Error) {
	return response, err
}

// Redis GET command.
func (client *client) GET(key []byte) ([]byte, os.Error) {
	return client.Send(GET, [][]byte{key})
}

// Redis MGET command.
func (client *client) MGET(keys [][]byte) ([]byte, os.Error) {
	return client.Send(MGET, keys)
}

// Redis PING command.
func (client *client) PING() ([]byte, os.Error) {
	return client.Send(PING, BlankArgs)
}

// Redis SET command.
func (client *client) SET(key []byte, value []byte) ([]byte, os.Error) {
	return client.Send(SET, [][]byte{key, value})
}

// Redis TTL command.
func (client *client) TTL() ([]byte, os.Error) {
	return client.Send(TTL, BlankArgs)
}

func (client *client) String() string {
	return fmt.Sprintf("<redis: %s>", client.address)
}

// The package is initialised with the maximum number of concurrent connections
// to a specific redis server set to 20. An application should change this if
// desired -- before making the first redis call.
func init() {
	MaxConnections = 20
	_ = bufio.NewReadWriter
	_ = net.Dial
}
