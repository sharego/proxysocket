package transport

const (
	// UnixName string
	UnixName = "unix"
)

// UnixConn is the Unix Domain Socket
type UnixConn struct {
	ProtoConnBase
	ServerAddr *net.UnixAddr
	RemoteAddr *net.UnixAddr
}

func init() {
	Register(UnixName)
}
