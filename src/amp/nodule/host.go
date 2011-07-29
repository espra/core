// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package nodule

import (
	"amp/logging"
	"amp/master"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

type Host struct {
	CachePath    string
	ControlTCP   net.Listener
	ControlUnix  net.Listener
	LastRef      uint64
	Listener     net.Listener
	Nodules      map[uint64]*Nodule
	ReadTimeout  int64
	WriteTimeout int64
}

func (host *Host) Run(debug bool) (err os.Error) {
	logging.InfoData("node", "foo")
	for {

	}
	return
}

func NewHost(runPath, hostAddress string, hostPort int, ctrlAddress string, ctrlPort int, initNodules string, nodulePaths []string, master *master.Client) (host *Host, err os.Error) {

	// Create the cache directory if it doesn't exist.
	cachePath := filepath.Join(runPath, "cache")
	err = os.MkdirAll(cachePath, 0755)
	if err != nil {
		return
	}

	ctrlTCP, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ctrlAddress, ctrlPort))
	if err != nil {
		return
	}

	socket := filepath.Join(runPath, "node.sock")
	_, err = os.Stat(socket)
	if err == nil {
		err = os.Remove(socket)
		if err != nil {
			return
		}
	}

	ctrlUnix, err := net.Listen("unix", socket)
	if err != nil {
		return
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", hostAddress, hostPort))
	if err != nil {
		return
	}

	host = &Host{
		CachePath:    cachePath,
		ControlTCP:   ctrlTCP,
		ControlUnix:  ctrlUnix,
		Listener:     listener,
		ReadTimeout:  60 * 1e9,
		WriteTimeout: 60 * 1e9,
	}

	return

}
