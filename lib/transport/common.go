package transport

import (
	"net"
	"time"
)

// Mode to indicate connection source
type Mode bool

const (
	//ConnectionIsServerAccept mode
	ConnectionIsServerAccept Mode = false
	//ConnectionIsDailerInitiator mode
	ConnectionIsDailerInitiator Mode = true
)

type ProtoConn interface {
	ProtoName() string
	Dail() error
	Serve(wg *sync.WaitGroup, ch chan <- *ProtoConn) error
}

// Proto is basic transport proto struct,
// should be embeding to other proto
type ProtoConnBase struct {
	Addr             string
	Name             string
	RawConn          *net.Conn
	Timeout          time.Duration
	IsConnectionless bool
	ConnMode         Mode
	IsClosed         bool
	// KeepAlive bool // Not Support Now, which request by top layer
}

func (p ProtoConnBase) ProtoName() string {
	return p.Name
}

var Protos map[string]interface
var lock = new(sync.Mutex)


func Register(name string){
	lock.Lock()
	Protos[name] = nil
	lock.Unlock()
}

type ProtoConnEvent struct{}

func (e *ProtoConnEvent) Dail() error {
	return errors.New("ProtoConnEvent can not dail")
}

func (e *ProtoConnEvent) Serve(wg *sync.WaitGroup, ch chan <- *ProtoConn) error{
	return errors.New("ProtoConnEvent can not serve")
}

type ProtoConnEventTimeout struct{
	ProtoConnEvent
	Timeout time.Duration
}

type ProtoConnEventClose struct{
	ProtoConnEvent
	Network string
	Addr string
}
