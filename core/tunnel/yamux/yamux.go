package yamux

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/libp2p/go-yamux/v5"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func init() {
	tunnel.RegisterTransport(tunnel.TransportYamux, NewTransport)
}

// NewTransport creates a new yamux transport
func NewTransport(opts *tunnel.TransportOptions) (tunnel.Transport, error) {
	return &yamuxTransport{opts: opts}, nil
}

type yamuxTransport struct {
	opts *tunnel.TransportOptions
}

func (t *yamuxTransport) Name() string {
	return tunnel.TransportYamux
}

func (t *yamuxTransport) Dial(ctx context.Context, addr string) (tunnel.Session, error) {
	dialer := &net.Dialer{Timeout: 30 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	if t.opts != nil && t.opts.EnableTLS {
		tlsConfig := &tls.Config{InsecureSkipVerify: t.opts.Insecure}
		conn = tls.Client(conn, tlsConfig)
	}

	session, err := yamux.Client(conn, t.buildConfig(), nil)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &yamuxSession{session: session, conn: conn}, nil
}

func (t *yamuxTransport) Listen(ctx context.Context, addr string) (tunnel.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	if t.opts != nil && t.opts.EnableTLS {
		cert, err := tls.LoadX509KeyPair(t.opts.CertFile, t.opts.KeyFile)
		if err != nil {
			ln.Close()
			return nil, err
		}
		ln = tls.NewListener(ln, &tls.Config{Certificates: []tls.Certificate{cert}})
	}

	return &yamuxListener{listener: ln, transport: t}, nil
}

func (t *yamuxTransport) buildConfig() *yamux.Config {
	cfg := yamux.DefaultConfig()
	if t.opts != nil {
		if t.opts.MaxStreams > 0 {
			cfg.AcceptBacklog = t.opts.MaxStreams
			cfg.MaxIncomingStreams = uint32(t.opts.MaxStreams)
		}
		if t.opts.KeepAliveInterval > 0 {
			cfg.KeepAliveInterval = time.Duration(t.opts.KeepAliveInterval) * time.Second
		}
		if t.opts.ConnectionWriteTimeout > 0 {
			cfg.ConnectionWriteTimeout = time.Duration(t.opts.ConnectionWriteTimeout) * time.Second
		}
	}
	return cfg
}

type yamuxListener struct {
	listener  net.Listener
	transport *yamuxTransport
}

func (l *yamuxListener) Accept() (tunnel.Session, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	session, err := yamux.Server(conn, l.transport.buildConfig(), nil)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &yamuxSession{session: session, conn: conn}, nil
}

func (l *yamuxListener) Close() error   { return l.listener.Close() }
func (l *yamuxListener) Addr() net.Addr { return l.listener.Addr() }

type yamuxSession struct {
	session *yamux.Session
	conn    net.Conn
	mu      sync.Mutex
}

// Open 打开一个新的流
func (s *yamuxSession) Open(ctx context.Context) (tunnel.Stream, error) {
	stream, err := s.session.Open(ctx)
	if err != nil {
		return nil, err
	}

	return &yamuxStream{stream: stream}, nil
}

// OpenWithPriority 打开指定优先级的流（1-10，1最高）
func (s *yamuxSession) OpenWithPriority(ctx context.Context, priority int) (tunnel.Stream, error) {
	// yamux 不支持优先级，直接调用 Open
	return s.Open(ctx)
}

func (s *yamuxSession) Accept() (tunnel.Stream, error) {
	stream, err := s.session.AcceptStream()
	if err != nil {
		return nil, err
	}
	return &yamuxStream{stream: stream}, nil
}

func (s *yamuxSession) Close() error         { return s.session.Close() }
func (s *yamuxSession) IsClosed() bool       { return s.session.IsClosed() }
func (s *yamuxSession) NumStreams() int      { return s.session.NumStreams() }
func (s *yamuxSession) LocalAddr() net.Addr  { return s.conn.LocalAddr() }
func (s *yamuxSession) RemoteAddr() net.Addr { return s.conn.RemoteAddr() }

type yamuxStream struct {
	stream net.Conn
}

func (s *yamuxStream) Read(p []byte) (int, error)         { return s.stream.Read(p) }
func (s *yamuxStream) Write(p []byte) (int, error)        { return s.stream.Write(p) }
func (s *yamuxStream) Close() error                       { return s.stream.Close() }
func (s *yamuxStream) LocalAddr() net.Addr                { return s.stream.LocalAddr() }
func (s *yamuxStream) RemoteAddr() net.Addr               { return s.stream.RemoteAddr() }
func (s *yamuxStream) SetDeadline(t time.Time) error      { return s.stream.SetDeadline(t) }
func (s *yamuxStream) SetReadDeadline(t time.Time) error  { return s.stream.SetReadDeadline(t) }
func (s *yamuxStream) SetWriteDeadline(t time.Time) error { return s.stream.SetWriteDeadline(t) }

// Priority 获取流优先级
func (s *yamuxStream) Priority() int {
	// yamux 不支持优先级，返回默认值
	return 5
}
