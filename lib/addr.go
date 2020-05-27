package lib

import (
	"errors"
	"net"
	"strings"
	"crypto/tls"
)

// ProxyProto Handle TCP,UDP,Unix Address
type ProxyProto struct {
	Addr           string
	IsTCP          bool
	IsUDP          bool
	IsUnix         bool
	TCPAddr        *net.TCPAddr
	UDPAddr        *net.UDPAddr
	UnixAddr       *net.UnixAddr
	ProtoPropeties interface{}
}

// TLSProtoPropeties using for TLS connection
type TLSProtoPropeties struct {
	Ca           tls.Certificate
	VerifyServer bool
	VerifyClient bool
}

// ResolveAddr parse like: tcp://10.0.0.1:8080
func ResolveAddr(protoaddr string) (pp *ProxyProto, err error) {
	network := "tcp"
	addr := ""

	if strings.HasPrefix(protoaddr, "//") {
		addr = protoaddr[2:]
	} else {
		parts := strings.SplitN(strings.TrimSpace(protoaddr), "://", 2)
		if len(parts) == 2 {
			if len(parts[0]) != 0 {
				network = parts[0]
			}
			addr = parts[1]
		} else {
			addr = parts[0]
		}
		if strings.HasPrefix(addr, "//") {
			addr = addr[2:]
		}
	}

	if len(network) == 0 || len(addr) == 0 {
		err = errors.New("invalid address: " + protoaddr)
		return nil, err
	}

	if strings.HasPrefix(network, "tcp") {
		a, err := net.ResolveTCPAddr(network, addr)
		if err != nil {
			return nil, err
		}
		pp = &ProxyProto{IsTCP: true, TCPAddr: a}
	} else if strings.HasPrefix(network, "udp") {
		a, err := net.ResolveUDPAddr(network, addr)
		if err != nil {
			return nil, err
		}
		pp = &ProxyProto{IsUDP: true, UDPAddr: a}
	} else if strings.HasPrefix(network, "unix") {
		a, err := net.ResolveUnixAddr(network, addr)
		if err != nil {
			return nil, err
		}
		pp = &ProxyProto{IsUnix: true, UnixAddr: a}
	} else if strings.HasPrefix(network, "tls") {
		a, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return nil, err
		}
		pp = &ProxyProto{
			TCPAddr: a,
			ProtoPropeties: &TLSProtoPropeties{
				Ca:           DefaultCa(),
				VerifyClient: true,
				VerifyServer: true,
			},
		}
	}
	pp.Addr = network + "://" + addr
	return pp, nil
}
