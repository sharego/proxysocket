package lib

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"fmt"
	"net"
	"strings"
	"crypto/x509"

	logging "github.com/go-fastlog/fastlog"
)

var log = logging.New(os.Stderr, "tunnel", logging.Ldebug)

// ForceCheckClient global setting force check client is authenticated
var ForceCheckClient string

// ForceCheckServer global setting force check server is authenticated
var ForceCheckServer string

// ForceCheckBoth global setting force check server and client is authenticated
var ForceCheckBoth string

// ServerConfig to config a app server
type ServerConfig struct {
	Name       string `json:"-"`
	In         string `json:"in"`
	Out        string `json:"out"`
	ServerCa         string `json:"serverca"`
	ClientCa         string `json:"clientca"`
	Incert     string `json:"incert"`
	Inkey      string `json:"inkey"`
	Inip string `json:"inip"`
	Nocheckin  bool   `json:"nocheckin"`
	Outcert    string `json:"outcert"`
	Outkey     string `json:"outkey"`
	Nocheckout bool   `json:"nocheckout"`
}

// ProxyChainTunnel compose TunnelServer and Dialer
type ProxyChainTunnel struct {
	InAddr string
	OutAddr string
	InProto *ProxyProto
	OutProto *ProxyProto

	s ProxyTunnelServer
	d ProxyTunnelDialer
}

// Serve A Tunnel Connect Inbound and Outbound
func (p *ProxyChainTunnel) Serve(sc *ServerConfig) {

	if e := p.ParseServerConfig(sc); e != nil {
		log.Fatal(e)
		return
	}

	inaddr := p.InProto
	outaddr := p.OutProto

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

// ParseServerConfig parse user config
func (p *ProxyChainTunnel) ParseServerConfig(sc *ServerConfig) (err error) {

	p.InAddr = sc.In
	p.OutAddr = sc.Out
	inaddr, err := ResolveAddr(p.InAddr)
	if err != nil {
		return fmt.Errorf("parse inbound address %s.In: %s, error: %s", sc.Name,p.InAddr, err)
	}

	outaddr, err := ResolveAddr(p.OutAddr)
	if err != nil {
		return fmt.Errorf("parse outbound address %s.Out: %s, error: %s", sc.Name, p.OutAddr, err)
	}

	if !inaddr.IsUDP && outaddr.IsUDP {
		return fmt.Errorf("not support create a tunnel from tls/tcp/unix to udp protocol, %s.In: %s, out: %s", sc.Name, inaddr.Addr, outaddr.Addr)
	} else if inaddr.IsUDP && !outaddr.IsUDP {
		log.Warnf("not support create a tunnel from tls/tcp/unix to udp protocol, %s.In: %s, out: %s", sc.Name, inaddr.Addr, outaddr.Addr)
	}

	p.InProto = inaddr
	p.OutProto = outaddr

	return p.parseTLSPropeties(sc)

}

func (p *ProxyChainTunnel) parseTLSPropeties(sc *ServerConfig) (err error) {

	defaultCa := DefaultCa()

	// using default ca to make server certificate if need

	if strings.HasPrefix(p.InProto.Addr, "tls") {
		var prop TLSProtoPropeties
		if len(sc.Incert) > 0 && len(sc.Inkey) > 0 {
			if cert, err := LoadCertKeyPair(sc.Incert, sc.Inkey); err == nil {
				prop.Cert = &cert
			} else {
				return fmt.Errorf("parse %s in error: %s", sc.Name, err)
			}
		} else if len(sc.Incert) > 0 {
			return fmt.Errorf("parse %s error: missing private key file path", sc.Name)
		} else if len(sc.Inkey) > 0{
			return fmt.Errorf("parse %s error: missing private key file path")
		} else {
			log.Warnf("parse %s, make server Certificate", sc.Name)
			if len(sc.Inip) > 0 {
				ip := net.ParseIP(sc.Inip)
				if ip == nil {
					return fmt.Errorf("parse %s Inip %s error", sc.Name, sc.Inip)
				}
			}
			cert, err := GenerateServerCert("localhost", sc.Inip, defaultCa)
			if err != nil {
				return fmt.Errorf("parse %s.clientca error: %s", sc.Name, err)
			}
			prop.Cert = cert
		}

		prop.VerifyClient = !sc.Nocheckin

		if len(ForceCheckClient) > 0 || len(ForceCheckBoth) > 0{
			log.Warnf("ignore config, %s force VerifyClient", sc.Name)
			prop.VerifyClient = true
		}

		if prop.VerifyClient {
			if strings.ToLower(sc.ClientCa) == "inner" {
				cert, _ := x509.ParseCertificate(defaultCa.Certificate[0])
				prop.Cacert = cert
			} else if len(sc.ClientCa) > 0 {
				if cert, err := LoadCert(sc.ClientCa); err == nil {
					prop.Cacert = cert
				} else {
					return err
				}
			} else {
				return fmt.Errorf("parse %s error: missing client ca", sc.Name)
			}
		}
		p.InProto.ProtoPropeties = &prop
	}

	if strings.HasPrefix(p.OutProto.Addr, "tls") {
		var prop TLSProtoPropeties
		if len(sc.Outcert) > 0 && len(sc.Outkey) > 0 {
			if cert, err := LoadCertKeyPair(sc.Outcert, sc.Outkey); err == nil {
				prop.Cert = &cert
			} else {
				return fmt.Errorf("parse %s out cert&key error: %s", sc.Name, err)
			}
		} else if len(sc.Outcert) > 0 {
			return fmt.Errorf("parse %s out error: missing private key file path", sc.Name)
		} else if len(sc.Outkey) > 0{
			return fmt.Errorf("parse %s out error: missing private key file path")
		} else {
			log.Warnf("parse %s, make client Certificate", sc.Name)
			cert, err := GenerateCert("localhost", defaultCa)
			if err != nil {
				return err
			}
			prop.Cert = cert
		}

		prop.VerifyServer = !sc.Nocheckout

		if len(ForceCheckServer) > 0 || len(ForceCheckBoth) > 0 {
			log.Warnf("ignore config, %s force VerifyServer", sc.Name)
			prop.VerifyServer = true
		}

		if prop.VerifyServer {
			if strings.ToLower(sc.ServerCa) == "inner" {
				cert, _ := x509.ParseCertificate(defaultCa.Certificate[0])
				prop.Cacert = cert
			} else if len(sc.ServerCa) > 0 {
				if cert, err := LoadCert(sc.ServerCa); err == nil {
					prop.Cacert = cert
				} else {
					return fmt.Errorf("parse %s.ServerCa error: %s", sc.Name, err)
				}
			} else {
				return fmt.Errorf("parse %s error: missing server ca", sc.Name)
			}
		}
		p.OutProto.ProtoPropeties = &prop
	}

	return nil
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
