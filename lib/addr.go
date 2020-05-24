package lib

import (
	"errors"
	"net"
	"strings"
)

// ProxyProtoAddr Handle TCP,UDP,Unix Address
type ProxyProtoAddr struct {
	Addr     string
	IsTCP    bool
	IsUDP    bool
	IsUnix   bool
	TCPAddr  *net.TCPAddr
	UDPAddr  *net.UDPAddr
	UnixAddr *net.UnixAddr
}

// ResolveAddr parse like: tcp://10.0.0.1:8080
func ResolveAddr(protoaddr string) (pa *ProxyProtoAddr, err error) {
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
		pa = &ProxyProtoAddr{IsTCP: true, TCPAddr: a}
	} else if strings.HasPrefix(network, "udp") {
		a, err := net.ResolveUDPAddr(network, addr)
		if err != nil {
			return nil, err
		}
		pa = &ProxyProtoAddr{IsUDP: true, UDPAddr: a}
	} else if strings.HasPrefix(network, "unix") {
		a, err := net.ResolveUnixAddr(network, addr)
		if err != nil {
			return nil, err
		}
		pa = &ProxyProtoAddr{IsUnix: true, UnixAddr: a}
	}
	pa.Addr = network + "://" + addr
	return pa, nil
}
