package quicconn

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"

	"github.com/pubgo/lava/v2/core/tunnel"
)

// WrapSession 将 quic-go 连接包装为 tunnel.Session。
func WrapSession(conn *quic.Conn) tunnel.Session {
	return &session{conn: conn}
}

type session struct {
	conn *quic.Conn
	mu   sync.Mutex
}

func (s *session) Open(ctx context.Context) (tunnel.Stream, error) {
	st, err := s.conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	return &stream{stream: st, conn: s.conn}, nil
}

func (s *session) OpenWithPriority(ctx context.Context, priority int) (tunnel.Stream, error) {
	return s.Open(ctx)
}

func (s *session) Accept() (tunnel.Stream, error) {
	st, err := s.conn.AcceptStream(context.Background())
	if err != nil {
		return nil, err
	}
	return &stream{stream: st, conn: s.conn}, nil
}

func (s *session) Close() error {
	return s.conn.CloseWithError(0, "closed")
}

func (s *session) IsClosed() bool {
	select {
	case <-s.conn.Context().Done():
		return true
	default:
		return false
	}
}

func (s *session) NumStreams() int { return -1 }

func (s *session) LocalAddr() net.Addr  { return s.conn.LocalAddr() }
func (s *session) RemoteAddr() net.Addr { return s.conn.RemoteAddr() }

type stream struct {
	stream *quic.Stream
	conn   *quic.Conn
}

func (s *stream) Read(p []byte) (int, error)  { return s.stream.Read(p) }
func (s *stream) Write(p []byte) (int, error) { return s.stream.Write(p) }
func (s *stream) Close() error                { return s.stream.Close() }
func (s *stream) LocalAddr() net.Addr         { return s.conn.LocalAddr() }
func (s *stream) RemoteAddr() net.Addr        { return s.conn.RemoteAddr() }

func (s *stream) SetDeadline(t time.Time) error {
	if err := s.stream.SetReadDeadline(t); err != nil {
		return err
	}
	return s.stream.SetWriteDeadline(t)
}

func (s *stream) SetReadDeadline(t time.Time) error  { return s.stream.SetReadDeadline(t) }
func (s *stream) SetWriteDeadline(t time.Time) error { return s.stream.SetWriteDeadline(t) }
func (s *stream) Priority() int                      { return 5 }

// WrapListener 包装 QUIC listener 为 tunnel.Listener。
func WrapListener(ln *quic.Listener, ctx context.Context) tunnel.Listener {
	return &listener{ln: ln, ctx: ctx}
}

type listener struct {
	ln  *quic.Listener
	ctx context.Context
}

func (l *listener) Accept() (tunnel.Session, error) {
	conn, err := l.ln.Accept(l.ctx)
	if err != nil {
		return nil, err
	}
	return WrapSession(conn), nil
}

func (l *listener) Close() error   { return l.ln.Close() }
func (l *listener) Addr() net.Addr { return l.ln.Addr() }

// DialSession ICE PacketConn 上发起 QUIC 并返回 tunnel.Session（带重试）。
func DialSession(
	ctx context.Context,
	pc net.PacketConn,
	remote net.Addr,
	opts *tunnel.TransportOptions,
) (tunnel.Session, error) {
	tr := &quic.Transport{Conn: pc}
	tlsConf := clientTLS(opts)
	cfg := quicConfig(opts)

	var lastErr error
	for attempt := 0; attempt < 12; attempt++ {
		conn, err := tr.Dial(ctx, remote, tlsConf, cfg)
		if err == nil {
			return WrapSession(conn), nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
	return nil, lastErr
}

// AcceptOne 在 ICE PacketConn 上监听并接受一条 QUIC 连接。
func AcceptOne(
	ctx context.Context,
	pc net.PacketConn,
	opts *tunnel.TransportOptions,
) (tunnel.Session, error) {
	ln, err := Listen(pc, opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = ln.Close() }()
	conn, err := ln.Accept(ctx)
	if err != nil {
		return nil, err
	}
	return WrapSession(conn), nil
}

// CopyStream 辅助测试：在 stream 间拷贝一次读写。
func CopyStream(w io.Writer, r io.Reader) (int64, error) {
	return io.Copy(w, r)
}
