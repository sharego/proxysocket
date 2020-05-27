package transport

const (
	// TLSName string
	TLSName = "tls"
)

// TLSConn is the security TCP
type TSLConn struct {
	ProtoConnBase
	ServerAddr *net.TCPAddr
	RemoteAddr *net.TCPAddr
}

func init() {
	Register(TLSName)
}
