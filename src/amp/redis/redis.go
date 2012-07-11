// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// The redis package provides a client library to interact with redis servers.
package redis

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	TcpConnection    = 0
	UnixConnection   = 1
	APPEND           = "APPEND"
	AUTH             = "AUTH"
	BGREWRITEAOF     = "BGREWRITEAOF"
	BGSAVE           = "BGSAVE"
	CONFIG           = "CONFIG"
	DBSIZE           = "DBSIZE"
	DECR             = "DECR"
	DECRBY           = "DECRBY"
	EXISTS           = "EXISTS"
	EXPIRE           = "EXPIRE"
	FLUSHALL         = "FLUSHALL"
	FLUSHDB          = "FLUSHDB"
	GET              = "GET"
	GETSET           = "GETSET"
	HDEL             = "HDEL"
	HEXISTS          = "HEXISTS"
	HGET             = "HGET"
	HGETALL          = "HGETALL"
	HINCRBY          = "HINCRBY"
	HKEYS            = "HKEYS"
	HLEN             = "HLEN"
	HMGET            = "HMGET"
	HMSET            = "HMSET"
	HSET             = "HSET"
	HVALS            = "HVALS"
	INCR             = "INCR"
	INCRBY           = "INCRBY"
	INFO             = "INFO"
	KEYS             = "KEYS"
	LASTSAVE         = "LASTSAVE"
	LINDEX           = "LINDEX"
	LLEN             = "LLEN"
	LPOP             = "LPOP"
	LPUSH            = "LPUSH"
	LRANGE           = "LRANGE"
	LREM             = "LREM"
	LSET             = "LSET"
	LTRIM            = "LTRIM"
	MGET             = "MGET"
	MOVE             = "MOVE"
	MSET             = "MSET"
	MSETNX           = "MSETNX"
	PING             = "PING"
	PSUBSCRIBE       = "PSUBSCRIBE"
	PUBLISH          = "PUBLISH"
	PUNSUBSCRIBE     = "PUNSUBSCRIBE"
	QUIT             = "QUIT"
	RANDOMKEY        = "RANDOMKEY"
	RENAME           = "RENAME"
	RENAMENX         = "RENAMENX"
	RPOP             = "RPOP"
	RPOPLPUSH        = "RPOPLPUSH"
	RPUSH            = "RPUSH"
	SADD             = "SADD"
	SAVE             = "SAVE"
	SCARD            = "SCARD"
	SDIFF            = "SDIFF"
	SDIFFSTORE       = "SDIFFSTORE"
	SELECT           = "SELECT"
	SET              = "SET"
	SETEX            = "SETEX"
	SETNX            = "SETNX"
	SHUTDOWN         = "SHUTDOWN"
	SINTER           = "SINTER"
	SINTERSTORE      = "SINTERSTORE"
	SISMEMBER        = "SISMEMBER"
	SLAVEOF          = "SLAVEOF"
	SMEMBERS         = "SMEMBERS"
	SMOVE            = "SMOVE"
	SORT             = "SORT"
	SPOP             = "SPOP"
	SRANDMEMBER      = "SRANDMEMBER"
	SREM             = "SREM"
	SUBSCRIBE        = "SUBSCRIBE"
	SUBSTR           = "SUBSTR"
	SUNION           = "SUNION"
	SUNIONSTORE      = "SUNIONSTORE"
	TTL              = "TTL"
	TYPE             = "TYPE"
	UNSUBSCRIBE      = "UNSUBSCRIBE"
	ZADD             = "ZADD"
	ZCARD            = "ZCARD"
	ZCOUNT           = "ZCOUNT"
	ZINCRBY          = "ZINCRBY"
	ZINTERSTORE      = "ZINTERSTORE"
	ZRANGE           = "ZRANGE"
	ZRANGEBYSCORE    = "ZRANGEBYSCORE"
	ZRANK            = "ZRANK"
	ZREM             = "ZREM"
	ZREMRANGEBYRANK  = "ZREMRANGEBYRANK"
	ZREMRANGEBYSCORE = "ZREMRANGEBYSCORE"
	ZREVRANGE        = "ZREVRANGE"
	ZREVRANK         = "ZREVRANK"
	ZSCORE           = "ZSCORE"
	ZUNIONSTORE      = "ZUNIONSTORE"
)

var (
	MaxConnections = 20
	connections    = make(map[string]chan *net.Conn, 100)
	connectionLock = make(chan int, 1)
	tickLock       = make(chan int, 1)
	CRLF           = []byte("\r\n")
)

var (
	TickRunning  bool  = false
	TickInterval int64 = 1 * (1 << 30)
	TickValue    int64 = 0
)

func Tick(interval int64) {
	if interval <= 0 {
		return
	}
	<-tickLock
	if TickRunning {
		tickLock <- 1
		return
	}
	TickRunning = true
	tickLock <- 1
	duration := time.Duration(interval)
	for TickRunning {
		TickValue = time.Now().UnixNano()
		<-time.After(duration)
	}
}

func StopTicking() {
	TickRunning = false
}

type KeyspaceProxy struct {
	Connected bool
	Servers   string
}

func Keyspace(servers string) *KeyspaceProxy {
	if !TickRunning {
		go Tick(TickInterval)
	}
	return &KeyspaceProxy{Servers: servers}
}

func (keyspace *KeyspaceProxy) Client(namespace string) *Redis {
	address := ""
	return &Redis{
		Address:        address,
		ConnectionType: TcpConnection,
		connection:     nil,
	}
}

func (keyspace *KeyspaceProxy) String() string {
	return fmt.Sprintf("<keyspace: %s>", keyspace.Servers)
}

type RedisError struct {
	Message string
}

func (err *RedisError) String() string {
	return "RedisError: " + string(err.Message)
}

type Redis struct {
	Address        string
	ConnectionType int
	multi          bool
	connected      bool
	connection     *net.Conn
}

// The ``Client`` constructor takes a single optional string parameter of the
// address of the redis server to connect to.
//
// If the address parameter is left out, it defaults to a TCP connection to
// ``localhost:6379``.
func Client(addr ...string) *Redis {

	var address string
	var connectionType int

	addrSlice := []string(addr)
	if len(addrSlice) > 0 {
		address = addrSlice[0]
		if strings.HasPrefix(address, "unix:") {
			address = address[5:]
			connectionType = UnixConnection
		}
	} else {
		address = "localhost:6379"
		connectionType = TcpConnection
	}

	return &Redis{
		Address:        address,
		ConnectionType: connectionType,
		connection:     nil,
	}

}

func (client *Redis) Connect() (err error) {

	if client.connected {
		return nil
	}

	if MaxConnections > 0 {
		pool, ok := connections[client.Address]
		if ok {
			conn := <-pool
			if conn != nil {
				fmt.Println("Connected to [cache]:", conn)
				client.connection = conn
				client.connected = true
				return
			}
		} else {
			<-connectionLock
			if _, ok = connections[client.Address]; !ok {
				pool := make(chan *net.Conn, MaxConnections)
				for i := 1; i < MaxConnections; i++ {
					pool <- nil
				}
				connections[client.Address] = pool
			}
			connectionLock <- 1
		}
	}

	var conn net.Conn

	if client.ConnectionType == TcpConnection {
		var la, ra *net.TCPAddr
		if ra, err = net.ResolveTCPAddr(client.Address); err != nil {
			return &net.OpError{"dial", "tcp " + client.Address, nil, err}
		}
		conn, err = net.DialTCP("tcp", la, ra)
		if err != nil {
			if MaxConnections > 0 {
				connections[client.Address] <- nil
			}
			return err
		}
	} else {
		var la *net.UnixAddr
		ra := &net.UnixAddr{client.Address, false}
		conn, err = net.DialUnix("unix", la, ra)
		if err != nil {
			if MaxConnections > 0 {
				connections[client.Address] <- nil
			}
			return err
		}
	}

	fmt.Println("Connected to:", conn)
	client.connection = &conn
	client.connected = true
	return nil

}

func (client *Redis) Close() {
	fmt.Println("Closing connection")
	if client.connected {
		client.connected = false
		if MaxConnections > 0 {
			if client.connection != nil {
				connections[client.Address] <- client.connection
			}
		} else {
			client.connection.Close()
		}
		client.connection = nil
	}
}

func (client *Redis) Send(command string, args ...[]byte) ([]byte, error) {

	err := client.Connect()
	if err != nil {
		return nil, err
	}

	request := bytes.NewBuffer(
		[]byte(fmt.Sprintf("*%d\r\n$%d\r\n%s\r\n", len(args)+1, len(command), command)))

	for _, arg := range args {
		request.Write([]byte(fmt.Sprintf("$%d\r\n", len(arg))))
		request.Write(arg)
		request.Write(CRLF)
	}

	_, err = client.connection.Write(request.Bytes())
	if err != nil {
		return nil, err
	}

	if !client.multi {
		client.Close()
	}

	return nil, err

}

// func (client *Client) sendCommand(cmd string, args []string) (data interface{}, err error) {
//         try:
//             cxn.write(''.join(request))
//         except socket.error:
//             self.close_connection()
//             raise
//         multi = txn = txn_end = persist = None

// Redis GET command.
func (client *Redis) GET(key []byte) ([]byte, error) {
	return client.Send(GET, key)
}

func (client *Redis) GETString(key string) ([]byte, error) {
	return client.Send(GET, []byte(key))
}

// Redis MGET command.
func (client *Redis) MGET(keys ...[]byte) ([]byte, error) {
	return client.Send(MGET, keys)
}

// Redis PING command.
func (client *Redis) PING() ([]byte, error) {
	return client.Send(PING)
}

// Redis SET command.
func (client *Redis) SET(key []byte, value []byte) ([]byte, error) {
	return client.Send(SET, key, value)
}

func (client *Redis) SETString(key string, value string) ([]byte, error) {
	return client.Send(SET, []byte(key), []byte(value))
}

// Redis TTL command.
func (client *Redis) TTL() ([]byte, error) {
	return client.Send(TTL)
}

func (client *Redis) String() string {
	return fmt.Sprintf("<redis: %s>", client.Address)
}

// The package is initialised with the maximum number of concurrent connections
// to an individual redis server set to 20. If an application wants to change
// this, it should do so before making the first redis client call.
func init() {
	connectionLock <- 1
	tickLock <- 1
	_ = bufio.NewReadWriter
	_ = net.Dial
}
