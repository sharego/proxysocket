package lib

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	logging "github.com/go-fastlog/fastlog"
)

var log = logging.New(os.Stderr, "tunnel", logging.Ldebug)

// ProxyChainTunnel compose TunnelServer and Dialer
type ProxyChainTunnel struct {
	InAddr        string
	OutAddr       string
	InProto   *ProxyProto
	OutProto *ProxyProto

	s ProxyTunnelServer
	d ProxyTunnelDialer
}

// Serve A Tunnel Connect Inbound and Outbound
func (p ProxyChainTunnel) Serve() {
	inaddr, err := ResolveAddr(p.InAddr)
	if err != nil {
		log.Errorf("parse inbound address %s, error: %s", p.InAddr, err)
		return
	}

	outaddr, err := ResolveAddr(p.OutAddr)
	if err != nil {
		log.Errorf("parse outbound address %s, error: %s", p.OutAddr, err)
		return
	}

	if !inaddr.IsUDP && outaddr.IsUDP {
		log.Errorf("not support create a tunnel from tls/tcp/unix to udp protocol, in: %s, out: %s", inaddr.Addr, outaddr.Addr)
		return
	} else if inaddr.IsUDP && !outaddr.IsUDP {
		log.Warnf("not support create a tunnel from tls/tcp/unix to udp protocol, in: %s, out: %s", inaddr.Addr, outaddr.Addr)
	}

	p.InProto = inaddr
	p.OutProto = outaddr

	var s ProxyTunnelServer

	if inaddr.IsTCP {
		s = NewProxyTunnelTCPServer()
	} else if inaddr.IsUDP {
		s = new(ProxyTunnelUDPServer)
	} else if inaddr.IsUnix {
		s = new(ProxyTunnelUnixServer)
	} else if inaddr.ProtoPropeties != nil {
		switch inaddr.ProtoPropeties.(type) {
		case *TLSProtoPropeties:
			s = new(ProxyTunnelTLSServer)
		default:
			log.Errorf("not support protocol, in: %s, out: %s", inaddr.Addr, outaddr.Addr)
		}
	} else {
		log.Errorf("not support protocol, in: %s, out: %s", inaddr.Addr, outaddr.Addr)
		return
	}

	wg := new(sync.WaitGroup)

	ch := s.Serve(inaddr, wg)

	if ch == nil {
		log.Errorf("create a %s server failed", inaddr.Addr)
		return
	}

	// If performace, use more goroutine here
	go p.HandleConnection(ch, wg)

	// wait server quit
	wg.Wait()

}

// HandleConnection start proxy data
func (p ProxyChainTunnel) HandleConnection(ch <-chan *ProxyChainConn, wg *sync.WaitGroup) {

	quitC := make(chan os.Signal, 1)
	signal.Notify(quitC, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	wg.Add(1)
	defer wg.Done()

	pwg := sync.WaitGroup{}

ProxyLabel:
	for {
		select {
		case <-quitC:
			break ProxyLabel
		case conn := <-ch:
			go func() {
				pwg.Add(1)
				defer pwg.Done()
				conn.Exchange(p.OutProto)
			}()
		default:
		}
	}

	pwg.Wait()

}
