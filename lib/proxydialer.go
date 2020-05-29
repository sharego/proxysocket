package lib

import (
	"errors"
	"net"
	"sync"
	"crypto/tls"
	"crypto/x509"
)

// DialerPools A Upstream Connection Pool
type DialerPools struct {
	mu         *sync.Mutex
	dialerpool map[*ProxyProto]ProxyTunnelDialer
}

// ProxyTunnelDialer a abstract of different proto client
type ProxyTunnelDialer interface {
	SetAddr(*ProxyProto)
	SupportMultiplex() bool
	// TCP is Connection-Oriented, UDP is Connectionless
	IsConnectionless() bool
	GetConn() (net.Conn, error)
	GetStream() (interface{}, error)
}

// AllDialerPools All DialerPools in Memory
var AllDialerPools = &DialerPools{
	mu:         new(sync.Mutex),
	dialerpool: make(map[*ProxyProto]ProxyTunnelDialer),
}

// GetDailer response a exits connection or create one
func (d *DialerPools) GetDailer(addr *ProxyProto) ProxyTunnelDialer {
	d.mu.Lock()
	if v, ok := d.dialerpool[addr]; ok {
		if v.SupportMultiplex() {
			return v
		}
	}
	d.mu.Unlock()

	var p ProxyTunnelDialer

	if addr.IsTCP {
		p = new(ProxyTunnelTCPDialer)
	} else if addr.IsUnix {
		p = new(ProxyTunnelUnixDialer)
	} else if addr.IsUDP {
		p = new(ProxyTunnelUDPDialer)
	} else if addr.ProtoPropeties != nil {
		switch v := addr.ProtoPropeties.(type) {
		case *TLSProtoPropeties:
			d := new(ProxyTunnelTLSDialer)
			d.TLSPropeties = v
			p = d
		default:
			return nil
		}
	} else {
		return nil
	}

	p.SetAddr(addr)

	d.mu.Lock()
	d.dialerpool[addr] = p
	d.mu.Unlock()

	return p
}

// TCP

// ProxyTunnelTCPDialer a tcp connection dailer
type ProxyTunnelTCPDialer struct {
	Addr *ProxyProto
}

// SupportMultiplex tcp dialer not support multiplex
func (p *ProxyTunnelTCPDialer) SupportMultiplex() bool {
	return false
}

// IsConnectionless tcp dialer is connection-oriented
func (p *ProxyTunnelTCPDialer) IsConnectionless() bool {
	return false
}

//SetAddr set a tcp ProxyProtoAddr
func (p *ProxyTunnelTCPDialer) SetAddr(a *ProxyProto) {
	p.Addr = a
}

// GetConn create a tcp connection
func (p *ProxyTunnelTCPDialer) GetConn() (net.Conn, error) {
	if p.Addr == nil {
		return nil, errors.New("not init dailer address")
	}
	addr := p.Addr.TCPAddr
	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// GetStream TCPDialer not support multiplex
func (p *ProxyTunnelTCPDialer) GetStream() (interface{}, error) {
	return nil, errors.New("Origin Unix not support multiplex")
}

// Unix

// ProxyTunnelUnixDialer a unix connection dailer
type ProxyTunnelUnixDialer struct {
	Addr *ProxyProto
}

// SupportMultiplex unix dialer not support multiplex
func (p *ProxyTunnelUnixDialer) SupportMultiplex() bool {
	return false
}

// IsConnectionless unix dialer is not connectionless
func (p *ProxyTunnelUnixDialer) IsConnectionless() bool {
	return false
}

//SetAddr set a unix ProxyProtoAddr
func (p *ProxyTunnelUnixDialer) SetAddr(a *ProxyProto) {
	p.Addr = a
}

// GetConn create a unix connection
func (p *ProxyTunnelUnixDialer) GetConn() (net.Conn, error) {
	if p.Addr == nil {
		return nil, errors.New("not init dailer address")
	}
	addr := p.Addr.UnixAddr
	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// GetStream unix not support multiplex
func (p *ProxyTunnelUnixDialer) GetStream() (interface{}, error) {
	return nil, errors.New("Origin TCP not support multiplex")
}

// UDP

// ProxyTunnelUDPDialer a udp connection dailer
type ProxyTunnelUDPDialer struct {
	Addr *ProxyProto
}

// SupportMultiplex udp dialer not support multiplex
func (p *ProxyTunnelUDPDialer) SupportMultiplex() bool {
	return false
}

// IsConnectionless udp dialer is connectionless
func (p *ProxyTunnelUDPDialer) IsConnectionless() bool {
	return true
}

//SetAddr set a tls ProxyProtoAddr
func (p *ProxyTunnelUDPDialer) SetAddr(a *ProxyProto) {
	p.Addr = a
}

// GetConn create a unix connection
func (p *ProxyTunnelUDPDialer) GetConn() (net.Conn, error) {
	if p.Addr == nil {
		return nil, errors.New("not init dailer address")
	}
	addr := p.Addr.UDPAddr
	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// GetStream udp not support multiplex
func (p *ProxyTunnelUDPDialer) GetStream() (interface{}, error) {
	return nil, errors.New("Origin UDP not support multiplex")
}

// TLS

// ProxyTunnelTLSDialer a udp connection dailer
type ProxyTunnelTLSDialer struct {
	Addr *ProxyProto
	TLSPropeties *TLSProtoPropeties
}

// SupportMultiplex udp dialer not support multiplex
func (p *ProxyTunnelTLSDialer) SupportMultiplex() bool {
	return false
}

// IsConnectionless udp dialer is connectionless
func (p *ProxyTunnelTLSDialer) IsConnectionless() bool {
	return true
}

//SetAddr set a udp ProxyProtoAddr
func (p *ProxyTunnelTLSDialer) SetAddr(a *ProxyProto) {
	p.Addr = a
}

// GetConn create a TLS connection
func (p *ProxyTunnelTLSDialer) GetConn() (net.Conn,error) {
	if p.Addr == nil {
		return nil, errors.New("not init dailer address")
	}

	if p.Addr.ProtoPropeties == nil {
		return nil, errors.New("not init tls dailer propeties")
	}

	pp := p.TLSPropeties

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*pp.Cert},
		InsecureSkipVerify: !pp.VerifyServer,
	}

	if pp.VerifyServer {
		pool := x509.NewCertPool()
		pool.AddCert(pp.Cacert)
		tlsConfig.RootCAs = pool
	}

	addr := p.Addr.TCPAddr

	conn, err := tls.Dial(addr.Network(), addr.String(), tlsConfig)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// GetStream TLS not support multiplex
func (p *ProxyTunnelTLSDialer) GetStream() (interface{}, error) {
	return nil, errors.New("Origin TLS not support multiplex")
}
