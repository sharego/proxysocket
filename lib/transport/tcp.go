package transport

import (
	"errors"
	"net"
	"sync"

	"github.com/coredns/coredns/plugin/pkg/log"
)

const (
	// TCPName string
	TCPName = "tcp"
)

func init() {
	Register(TCPName)
}

// TCPConn is the Origin TCP/IP Layer
type TCPConn struct {
	ProtoConnBase
	ServerAddr *net.TCPAddr
	RemoteAddr *net.TCPAddr
}

// NewTCPConn construct a new TCP
func NewTCPConn() *TCPConn {
	t := TCP{
		Name:             TCPName,
		IsConnectionless: false,
	}
	// TODO
	return t
}

// Dail connect RemoteAddr
func (t *TCPConn) Dail() (err error) {
	if t.RemoteAddr == nil {
		return errors.New("not init dailer address")
	}
	addr := t.RemoteAddr
	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return
	}
	return
}

func (t *TCPConn) Serve(wg *sync.WaitGroup, ch chan<- *ProtoConn) (err error) {
	if t.ServerAddr == nil {
		return errors.New("not init listen address")
	}
	addr := t.ServerAddr
	listener, err := net.ListenTCP(addr.Network(), addr)
	if err != nil {
		log.Errorf("create tcp socket listen on %s failed: %s", addr.Addr, err)
		return nil
	}

}
