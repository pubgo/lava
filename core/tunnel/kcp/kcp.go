package kcp

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func init() {
	tunnel.RegisterTransport(tunnel.TransportKCP, NewTransport)
}

// NewTransport creates a new KCP transport
func NewTransport(opts *tunnel.TransportOptions) (tunnel.Transport, error) {
	return &kcpTransport{opts: opts}, nil
}

type kcpTransport struct {
	opts *tunnel.TransportOptions
}

func (t *kcpTransport) Name() string {
	return tunnel.TransportKCP
}

func (t *kcpTransport) Dial(ctx context.Context, addr string) (tunnel.Session, error) {
	// 创建 KCP 连接
	conn, err := kcp.DialWithOptions(addr, nil, 10, 3)
	if err != nil {
		return nil, err
	}

	// 配置 KCP 参数
	t.configureKCPConn(conn)

	var netConn net.Conn = conn
	if t.opts != nil && t.opts.EnableTLS {
		tlsConfig := tunnel.BuildClientTLSConfig(t.opts)
		netConn = tls.Client(conn, tlsConfig)
	}

	// 使用 smux 进行多路复用
	smuxConfig := t.buildSmuxConfig()
	session, err := smux.Client(netConn, smuxConfig)
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			return nil, closeErr
		}
		return nil, err
	}

	return &kcpSession{
		session:   session,
		conn:      conn,
		tlsConn:   netConn,
		transport: t,
	}, nil
}

func (t *kcpTransport) Listen(ctx context.Context, addr string) (tunnel.Listener, error) {
	ln, err := kcp.ListenWithOptions(addr, nil, 10, 3)
	if err != nil {
		return nil, err
	}

	return &kcpListener{
		listener:  ln,
		transport: t,
		ctx:       ctx,
	}, nil
}

func (t *kcpTransport) configureKCPConn(conn *kcp.UDPSession) {
	// 设置 KCP 模式：快速模式
	// nodelay: 0 关闭, 1 开启
	// interval: 内部更新间隔 ms
	// resend: 快速重传模式，0 关闭，可以设置为 2（2 次 ACK 跨越将会直接重传）
	// nc: 是否关闭流控，0 不关闭，1 关闭
	conn.SetNoDelay(1, 10, 2, 1)

	// 设置窗口大小
	conn.SetWindowSize(256, 256)

	// 设置读写缓冲区
	if err := conn.SetReadBuffer(4 * 1024 * 1024); err != nil {
		_ = err // ignore non-fatal buffer errors
	}
	if err := conn.SetWriteBuffer(4 * 1024 * 1024); err != nil {
		_ = err // ignore non-fatal buffer errors
	}

	// 设置 MTU
	conn.SetMtu(1350)

	// 设置 ACK 无延迟
	conn.SetACKNoDelay(true)
}

func (t *kcpTransport) buildSmuxConfig() *smux.Config {
	cfg := smux.DefaultConfig()
	if t.opts != nil {
		if t.opts.MaxStreams > 0 {
			cfg.MaxReceiveBuffer = t.opts.MaxStreams * 65536
			cfg.MaxStreamBuffer = 65536
		}
		if t.opts.KeepAliveInterval > 0 {
			cfg.KeepAliveInterval = time.Duration(t.opts.KeepAliveInterval) * time.Second
		}
	}
	return cfg
}

type kcpListener struct {
	listener  *kcp.Listener
	transport *kcpTransport
	ctx       context.Context
}

func (l *kcpListener) Accept() (tunnel.Session, error) {
	conn, err := l.listener.AcceptKCP()
	if err != nil {
		return nil, err
	}

	// 配置 KCP 参数
	l.transport.configureKCPConn(conn)

	var netConn net.Conn = conn
	if l.transport.opts != nil && l.transport.opts.EnableTLS {
		tlsConfig, err := tunnel.BuildServerTLSConfig(l.transport.opts)
		if err != nil {
			if closeErr := conn.Close(); closeErr != nil {
				return nil, closeErr
			}
			return nil, err
		}
		netConn = tls.Server(conn, tlsConfig)
	}

	// 使用 smux 进行多路复用
	smuxConfig := l.transport.buildSmuxConfig()
	session, err := smux.Server(netConn, smuxConfig)
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			return nil, closeErr
		}
		return nil, err
	}

	return &kcpSession{
		session:   session,
		conn:      conn,
		tlsConn:   netConn,
		transport: l.transport,
	}, nil
}

func (l *kcpListener) Close() error   { return l.listener.Close() }
func (l *kcpListener) Addr() net.Addr { return l.listener.Addr() }

type kcpSession struct {
	session   *smux.Session
	conn      *kcp.UDPSession
	tlsConn   net.Conn
	transport *kcpTransport
	mu        sync.Mutex
	closed    atomic.Bool
}

func (s *kcpSession) Open(ctx context.Context) (tunnel.Stream, error) {
	stream, err := s.session.OpenStream()
	if err != nil {
		return nil, err
	}
	return &kcpStream{stream: stream, session: s}, nil
}

// OpenWithPriority 打开指定优先级的流（1-10，1最高）
func (s *kcpSession) OpenWithPriority(ctx context.Context, priority int) (tunnel.Stream, error) {
	// kcp 不支持优先级，直接调用 Open
	return s.Open(ctx)
}

func (s *kcpSession) Accept() (tunnel.Stream, error) {
	stream, err := s.session.AcceptStream()
	if err != nil {
		return nil, err
	}
	return &kcpStream{stream: stream, session: s}, nil
}

func (s *kcpSession) Close() error {
	if s.closed.Swap(true) {
		return nil
	}
	if err := s.session.Close(); err != nil {
		return err
	}
	return s.conn.Close()
}

func (s *kcpSession) IsClosed() bool {
	return s.session.IsClosed()
}

func (s *kcpSession) NumStreams() int {
	return s.session.NumStreams()
}

func (s *kcpSession) LocalAddr() net.Addr  { return s.conn.LocalAddr() }
func (s *kcpSession) RemoteAddr() net.Addr { return s.conn.RemoteAddr() }

type kcpStream struct {
	stream  *smux.Stream
	session *kcpSession
}

func (s *kcpStream) Read(p []byte) (int, error)  { return s.stream.Read(p) }
func (s *kcpStream) Write(p []byte) (int, error) { return s.stream.Write(p) }
func (s *kcpStream) Close() error                { return s.stream.Close() }
func (s *kcpStream) LocalAddr() net.Addr         { return s.session.conn.LocalAddr() }
func (s *kcpStream) RemoteAddr() net.Addr        { return s.session.conn.RemoteAddr() }

func (s *kcpStream) SetDeadline(t time.Time) error {
	if err := s.stream.SetReadDeadline(t); err != nil {
		return err
	}
	return s.stream.SetWriteDeadline(t)
}

func (s *kcpStream) SetReadDeadline(t time.Time) error  { return s.stream.SetReadDeadline(t) }
func (s *kcpStream) SetWriteDeadline(t time.Time) error { return s.stream.SetWriteDeadline(t) }

// Priority 获取流优先级
func (s *kcpStream) Priority() int {
	// kcp 不支持优先级，返回默认值
	return 5
}
