package lib

import (
	"io"
	"net"
	"sync"
	"strings"
	"time"
)

// ProxyChainConn a concrate inbound and outbound connection pair
type ProxyChainConn struct {
	inConn          net.Conn
	InUDPRemoteAddr *net.UDPAddr
	UDPData         []byte
	outConn         net.Conn
	IsClosed        bool
}

// Exchange on Connection or NoConnection

// connection 1 : Client <-1-> ProxyServer
// connection 2 : ProxyDailer <-2-> UpstreamService
// connection-orintend: Duplex/Linked
// connectionless: Simplex

// |    1    |    2    |   Things    |
// |---------|---------|-------------|
// | Linked  | Linked  |  Sync Break |
// | Linked  | Simplex | Not Support |
// | Simplex | Linked  |  Time Break |
// | Simplex | Simplex |   No Need   |

// Exchange inbound and outbound connection data
func (c *ProxyChainConn) Exchange(to *ProxyProtoAddr) {
	if c == nil {
		return
	}

	dailer := AllDialerPools.GetDailer(to)

	if dailer == nil {
		log.Errorf("get a dailer to %s failed", to.Addr)
		c.Close()
		return
	}

	if dailer.SupportMultiplex() {
		log.Infof("support multiplex")
	}

	if conn, err := dailer.GetConn(); err == nil {
		c.outConn = conn
	} else {
		log.Errorf("connect %s failed: %s", to.Addr, err)
		c.Close()
		return
	}

	// on UDP Server Mode
	if c.InUDPRemoteAddr != nil {
		// send request to proxy service by dialer
		conn := c.outConn
		n, err := conn.Write(c.UDPData)
		if n != len(c.UDPData) || err != nil {
			log.Errorf("send %d bytes to %s error", len(c.UDPData), to.Addr)
			c.Close()
			return
		}

		// receive data response
		conn.SetReadDeadline(time.Now().Add(time.Second * 3))
		buf := make([]byte, 1500)
		readSize, err := conn.Read(buf)

		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && !opErr.Timeout() {
				// error yet, shoule response back to client?
				log.Errorf("read data(%d done) from %s error: %v", readSize, to.Addr, err)
			}
		}

		// Write response back
		writeSize, err := c.inConn.(*net.UDPConn).WriteToUDP(buf[:readSize], c.InUDPRemoteAddr)
		if writeSize != readSize || err != nil {
			log.Errorf("write %d bytes(%d done) to %s, error: %v", readSize, writeSize, c.InUDPRemoteAddr.String(), err)
		}

		c.Close()
		return
	}

	wg := sync.WaitGroup{}

	inConnClosed := false
	outConnClosed := false

	cp := func(src, dst net.Conn) {
		defer wg.Done()
		// buf := make([]byte, 16*1024) // 16 KB
		if !c.IsClosed {

			// On the paire connection any is Closed
			// There is no need to wait recieve more data and copy to the other
			timeoutCount := 0
			var totalSize int64 = 0
			for !inConnClosed && !outConnClosed {

				src.SetReadDeadline(time.Now().Add(time.Second * 3))

				// Reader From src, Write to dst
				size, err := io.CopyBuffer(dst, src, nil)
				totalSize += size
				if err != nil {
					if opErr, ok := err.(*net.OpError); ok {
						if opErr.Timeout() {
							timeoutCount++
						} else {
							log.Errorf("read data(%d done) from %s, opError: %v", totalSize, to.Addr, *opErr)
							// If error, then close all connection
							if !opErr.Temporary() {
								inConnClosed = true
							}
						}
					} else {
						// error yet, shoule response back to client?
						log.Errorf("read data(%d done) from %s, error: %v", totalSize, to.Addr, err)
						inConnClosed = true
					}
				} else {
					timeoutCount = 0
				}

				if timeoutCount == 0 {
					if src == c.inConn {
						inConnClosed = true
					} else {
						outConnClosed = true
					}
				}

				// It should be long for some tcp using keep-alive and no heartbeat data
				if timeoutCount >= 600 { // 30 minutes
					log.Infof("check read from unix to tcp, limit unix data input time of %d minutes", 30)
					break
				}
			}
			log.Infof("transfor %d bytes from %s to %s", totalSize, src.RemoteAddr(), dst.RemoteAddr())

		}
	}

	log.Infof("tunnel opened %s <-> [%s, %s] <-> %s", c.inConn.RemoteAddr(), c.inConn.LocalAddr(), c.outConn.LocalAddr(), c.outConn.RemoteAddr())

	// transfer data

	wg.Add(2)
	// proxy request from inbound to outbound
	go cp(c.inConn, c.outConn)
	// proxy response from outbound to inbound
	go cp(c.outConn, c.inConn)

	// Block no timeout
	wg.Wait()

	if !c.IsClosed {
		if c.inConn == nil {
			log.Infof("closing connection pair: nil <-> [nil, %s] <-> %s", c.outConn.LocalAddr(), c.outConn.RemoteAddr())
		} else if c.outConn == nil {
			log.Infof("closing connection pair: %s <-> [%s, nil] <-> nil", c.inConn.RemoteAddr(), c.inConn.LocalAddr())
		} else {
			log.Infof("closing connection pair: %s <-> [%s, %s] <-> %s", c.inConn.RemoteAddr(), c.inConn.LocalAddr(), c.outConn.LocalAddr(), c.outConn.RemoteAddr())
		}
	}

	// When to support multiplex, cannot close here
	c.Close()

}

// Close connection pair
func (c *ProxyChainConn) Close() {
	if c == nil {
		return
	}
	if c.IsClosed {
		return
	}
	// UDP Inbound Connection is a udp server, chould not be closed here
	if c.inConn != nil && !strings.HasPrefix(c.inConn.LocalAddr().Network(), "udp") {
		c.inConn.Close()
		c.inConn = nil
	}

	// If want to build a connection for tcp <-> udp <-> udp <-> tcp
	// Cannot close this connection, so not implements them
	if c.outConn != nil {
		c.outConn.Close()
		c.outConn = nil
	}
	c.IsClosed = true
}
