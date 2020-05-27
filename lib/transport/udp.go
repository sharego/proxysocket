package transport

const (
	// UDPName string
	UDPName = "udp"
)

// UDPConn is the Origin TCP/IP Layer
type UDPConn struct {
	ProtoConnBase
	ServerAddr *net.TCPAddr
	RemoteAddr *net.TCPAddr
	CarryData        []byte
}

func init() {
	Register(UDPName)
}
