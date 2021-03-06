package lib

import (
	"container/list"

	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ProxyTunnelServer abstract of different proto server
type ProxyTunnelServer interface {
	// Listen on addr, tel main goruntine when finish by wg
	Serve(addr *ProxyProtoAddr, wg *sync.WaitGroup) chan *ProxyChainConn
}

// ProxyTunnelTCPServer a tcp tunnel server
type ProxyTunnelTCPServer struct {
	mu    *sync.Mutex
	conns *list.List
}

// NewProxyTunnelTCPServer new TCPServer and set Propreties
func NewProxyTunnelTCPServer() ProxyTunnelServer {
	s := new(ProxyTunnelTCPServer)
	s.mu = new(sync.Mutex)
	s.conns = list.New()
	return s
}

// ProxyTunnelUDPServer a udp tunnel server
type ProxyTunnelUDPServer struct {
	Addr *net.UDPAddr
}

// ProxyTunnelUnixServer a unix tunnel server
type ProxyTunnelUnixServer struct {
	Addr *net.UnixAddr
}

// Serve a tcp listenner
func (s ProxyTunnelTCPServer) Serve(addr *ProxyProtoAddr, wg *sync.WaitGroup) chan *ProxyChainConn {
	listener, err := net.ListenTCP(addr.TCPAddr.Network(), addr.TCPAddr)
	if err != nil {
		log.Errorf("create tcp socket listen on %s failed: %s", addr.Addr, err)
		return nil
	}

	ch := make(chan *ProxyChainConn)

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Infof("start a server listen on %s, waiting to accept connection", addr.Addr)

		quitC := make(chan os.Signal, 1)
		signal.Notify(quitC, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		timeoutCount := 0
	AcceptLoop:
		for {
			select {
			case <-quitC:
				listener.Close()
				s.mu.Lock()
				for e := s.conns.Front(); e != nil; e = e.Next() {
					c := e.Value.(*ProxyChainConn)
					c.Close()
				}
				s.mu.Unlock()
				break AcceptLoop
			default:
			}
			listener.SetDeadline(time.Now().Add(1 * time.Second))
			conn, err := listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					timeoutCount++
					if timeoutCount < 3 {
						continue
					} else {
						s.mu.Lock()
						for e := s.conns.Front(); e != nil; e = e.Next() {
							c := e.Value.(*ProxyChainConn)
							if c.IsClosed {
								s.conns.Remove(e)
							}
						}
						s.mu.Unlock()
					}
				} else {
					log.Error(err)
				}
			} else {
				timeoutCount = 0
				log.Infof("accept a connection: %s -> %s", conn.RemoteAddr().String(), conn.LocalAddr().String())
				c := &ProxyChainConn{inConn: conn}
				ch <- c
				s.mu.Lock()
				s.conns.PushBack(c)
				s.mu.Unlock()
			}
		}

	}()

	return ch

}

// Serve a udp listenner
func (s ProxyTunnelUDPServer) Serve(addr *ProxyProtoAddr, wg *sync.WaitGroup) chan *ProxyChainConn {
	conn, err := net.ListenUDP(addr.UDPAddr.Network(), addr.UDPAddr)
	if err != nil {
		log.Errorf("create udp socket listen on %s failed: %s", addr.Addr, err)
		return nil
	}

	ch := make(chan *ProxyChainConn)

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Infof("start a server listen on %s, waiting to accept connection", addr.Addr)

		quitC := make(chan os.Signal, 1)
		signal.Notify(quitC, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	ConnLoop:
		for {
			select {
			case <-quitC:
				conn.Close()
				break ConnLoop
			default:
			}

			buf := make([]byte, 1500) // 1500 is Max UDP Packet Size

			conn.SetDeadline(time.Now().Add(3 * time.Second))
			size, remoteAddr, err := conn.ReadFromUDP(buf)

			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && !opErr.Timeout() {
					log.Error(err)
					// What should todo here?
				}
			}

			if size == 0 {
				continue
			}

			log.Infof("receive udp %d bytes from %s on %s", size, remoteAddr.String(), conn.LocalAddr().String())

			c := &ProxyChainConn{
				inConn:          conn,
				InUDPRemoteAddr: remoteAddr,
				UDPData:         buf[:size],
			}
			ch <- c
		}
	}()

	return ch

}

// Serve a unix listenner
func (s ProxyTunnelUnixServer) Serve(addr *ProxyProtoAddr, wg *sync.WaitGroup) chan *ProxyChainConn {
	listener, err := net.ListenUnix(addr.UnixAddr.Network(), addr.UnixAddr)
	if err != nil {
		log.Errorf("create unix socket listen on %s failed: %s", addr.Addr, err)
		return nil
	}

	ch := make(chan *ProxyChainConn)
	wg.Add(1)

	go func() {
		defer wg.Done()

		log.Infof("start a server listen on %s, waiting to accept connection", addr.Addr)

		quitC := make(chan os.Signal, 1)
		signal.Notify(quitC, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	AcceptLoop:
		for {
			select {
			case <-quitC:
				listener.Close()
				break AcceptLoop
			default:
			}

			listener.SetDeadline(time.Now().Add(1 * time.Second))
			conn, err := listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				log.Error(err)
			} else {
				log.Infof("accept a connection: %s -> %s", conn.RemoteAddr().String(), conn.LocalAddr().String())
				c := &ProxyChainConn{inConn: conn}
				ch <- c
			}
		}

		// After Unix Server Close, Should Remove sock file
		if err := os.Remove(addr.UnixAddr.String()); !os.IsNotExist(err) {
			log.Errorf("Remove file: %s, failed: %s", addr.UnixAddr.String(), err)
		}

	}()

	return ch
}
