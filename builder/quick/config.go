package quick

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/logging"
	"github.com/lucas-clemente/quic-go/quictrace"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
)

type Cfg struct {
	Versions                              []quic.VersionNumber
	ConnectionIDLength                    int
	HandshakeTimeout                      time.Duration
	MaxIdleTimeout                        time.Duration
	MaxReceiveStreamFlowControlWindow     uint64
	MaxReceiveConnectionFlowControlWindow uint64
	MaxIncomingStreams                    int64
	MaxIncomingUniStreams                 int64
	StatelessResetKey                     []byte
	KeepAlive                             bool
	TokenStore                            quic.TokenStore
	QuicTracer                            quictrace.Tracer
	Tracer                                logging.Tracer
	AcceptToken                           func(clientAddr net.Addr, token *quic.Token) bool
}

func (t Cfg) ListenAddr(addr string, tlsConf *tls.Config) (quic.Listener, error) {
	ln, err := quic.ListenAddr(addr, tlsConf, t.ToCfg())
	return ln, xerror.Wrap(err)
}

// ListenConn creates a QUIC listener on the given network interface
func (t Cfg) ListenConn(addr string, tlsConf *tls.Config) (net.Listener, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, &net.OpError{Op: "listen", Net: "udp", Source: nil, Addr: nil, Err: err}
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	ln, err := quic.Listen(conn, tlsConf, t.ToCfg())
	if err != nil {
		return nil, err
	}

	return &listener{conn: conn, server: ln}, nil
}

func (t Cfg) DialAddr(addr string, tlsConf *tls.Config) (quic.Session, error) {
	session, err := quic.DialAddr(addr, tlsConf, t.ToCfg())
	return session, xerror.Wrap(err)
}

// DialConn creates a new QUIC connection
// it returns once the connection is established and secured with forward-secure keys
func (t Cfg) DialConn(addr string, tlsConf *tls.Config) (_ net.Conn, err error) {
	defer xerror.RespErr(&err)

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	xerror.Panic(err)

	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	xerror.Panic(err)

	// DialAddr returns once a forward-secure connection is established
	session, err := quic.Dial(udpConn, udpAddr, addr, tlsConf, t.ToCfg())
	xerror.Panic(err)

	stream, err := session.OpenStreamSync(context.Background())
	xerror.Panic(err)

	return &conn{conn: udpConn, session: session, stream: stream}, nil
}

func (t Cfg) ToCfg() *quic.Config {
	var cfg = quic.Config{}

	xerror.Panic(merge.CopyStruct(&cfg, &t))

	return &cfg
}

func DefaultCfg() Cfg {
	return Cfg{}
}
