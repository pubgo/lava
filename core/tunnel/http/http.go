package http

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func init() {
	tunnel.RegisterTransport(tunnel.TransportHTTP, NewTransport)
}

// NewTransport creates a new HTTP CONNECT transport
func NewTransport(opts *tunnel.TransportOptions) (tunnel.Transport, error) {
	return &httpTransport{opts: opts}, nil
}

type httpTransport struct {
	opts *tunnel.TransportOptions
}

func (t *httpTransport) Name() string {
	return tunnel.TransportHTTP
}

func (t *httpTransport) Dial(ctx context.Context, addr string) (tunnel.Session, error) {
	// HTTP CONNECT 方式建立隧道
	dialer := &net.Dialer{Timeout: 30 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	if t.opts != nil && t.opts.EnableTLS {
		tlsConfig := &tls.Config{InsecureSkipVerify: t.opts.Insecure}
		conn = tls.Client(conn, tlsConfig)
	}

	// 发送 CONNECT 请求建立隧道
	req := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: make(http.Header),
	}
	req.Header.Set("Proxy-Connection", "keep-alive")
	req.Header.Set("X-Tunnel-Protocol", "lava-tunnel/1.0")

	if err := req.Write(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send CONNECT request: %w", err)
	}

	// 读取响应
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read CONNECT response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		conn.Close()
		return nil, fmt.Errorf("CONNECT request failed: %s", resp.Status)
	}

	return &httpSession{
		conn:     conn,
		isServer: false,
		streamID: 0,
	}, nil
}

func (t *httpTransport) Listen(ctx context.Context, addr string) (tunnel.Listener, error) {
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

	return &httpListener{
		listener:  ln,
		transport: t,
		ctx:       ctx,
	}, nil
}

type httpListener struct {
	listener  net.Listener
	transport *httpTransport
	ctx       context.Context
}

func (l *httpListener) Accept() (tunnel.Session, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}

	// 处理 CONNECT 请求
	br := bufio.NewReader(conn)
	req, err := http.ReadRequest(br)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read request: %w", err)
	}

	if req.Method != "CONNECT" {
		resp := &http.Response{
			StatusCode: http.StatusMethodNotAllowed,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header),
		}
		resp.Write(conn)
		conn.Close()
		return nil, fmt.Errorf("expected CONNECT method, got %s", req.Method)
	}

	// 发送 200 响应
	resp := &http.Response{
		StatusCode: http.StatusOK,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}
	resp.Header.Set("X-Tunnel-Protocol", "lava-tunnel/1.0")

	if err := resp.Write(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to write response: %w", err)
	}

	return &httpSession{
		conn:     conn,
		isServer: true,
		streamID: 0,
	}, nil
}

func (l *httpListener) Close() error   { return l.listener.Close() }
func (l *httpListener) Addr() net.Addr { return l.listener.Addr() }

// httpSession HTTP 隧道会话
// 注意：HTTP CONNECT 本身不支持多路复用，这里通过简单的帧协议模拟
type httpSession struct {
	conn     net.Conn
	isServer bool
	streamID uint32
	mu       sync.Mutex
	closed   atomic.Bool
	streams  sync.Map // streamID -> *httpStream
}

func (s *httpSession) Open(ctx context.Context) (tunnel.Stream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed.Load() {
		return nil, fmt.Errorf("session is closed")
	}

	id := atomic.AddUint32(&s.streamID, 1)
	stream := &httpStream{
		session:  s,
		streamID: id,
		readBuf:  make(chan []byte, 16),
		done:     make(chan struct{}),
	}
	s.streams.Store(id, stream)

	return stream, nil
}

func (s *httpSession) Accept() (tunnel.Stream, error) {
	// 对于简单的 HTTP CONNECT，只返回一个基于底层连接的流
	if s.closed.Load() {
		return nil, fmt.Errorf("session is closed")
	}

	return &httpDirectStream{
		conn:    s.conn,
		session: s,
	}, nil
}

func (s *httpSession) Close() error {
	if s.closed.Swap(true) {
		return nil
	}
	return s.conn.Close()
}

func (s *httpSession) IsClosed() bool {
	return s.closed.Load()
}

func (s *httpSession) NumStreams() int {
	count := 0
	s.streams.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

func (s *httpSession) LocalAddr() net.Addr  { return s.conn.LocalAddr() }
func (s *httpSession) RemoteAddr() net.Addr { return s.conn.RemoteAddr() }

// httpStream 多路复用流（基于帧协议）
type httpStream struct {
	session  *httpSession
	streamID uint32
	readBuf  chan []byte
	done     chan struct{}
	closed   atomic.Bool
}

func (s *httpStream) Read(p []byte) (int, error) {
	select {
	case data := <-s.readBuf:
		return copy(p, data), nil
	case <-s.done:
		return 0, io.EOF
	}
}

func (s *httpStream) Write(p []byte) (int, error) {
	if s.closed.Load() {
		return 0, fmt.Errorf("stream closed")
	}
	return s.session.conn.Write(p)
}

func (s *httpStream) Close() error {
	if s.closed.Swap(true) {
		return nil
	}
	close(s.done)
	s.session.streams.Delete(s.streamID)
	return nil
}

func (s *httpStream) LocalAddr() net.Addr                { return s.session.conn.LocalAddr() }
func (s *httpStream) RemoteAddr() net.Addr               { return s.session.conn.RemoteAddr() }
func (s *httpStream) SetDeadline(t time.Time) error      { return s.session.conn.SetDeadline(t) }
func (s *httpStream) SetReadDeadline(t time.Time) error  { return s.session.conn.SetReadDeadline(t) }
func (s *httpStream) SetWriteDeadline(t time.Time) error { return s.session.conn.SetWriteDeadline(t) }

// httpDirectStream 直接使用底层连接的流
type httpDirectStream struct {
	conn    net.Conn
	session *httpSession
}

func (s *httpDirectStream) Read(p []byte) (int, error)         { return s.conn.Read(p) }
func (s *httpDirectStream) Write(p []byte) (int, error)        { return s.conn.Write(p) }
func (s *httpDirectStream) Close() error                       { return s.session.Close() }
func (s *httpDirectStream) LocalAddr() net.Addr                { return s.conn.LocalAddr() }
func (s *httpDirectStream) RemoteAddr() net.Addr               { return s.conn.RemoteAddr() }
func (s *httpDirectStream) SetDeadline(t time.Time) error      { return s.conn.SetDeadline(t) }
func (s *httpDirectStream) SetReadDeadline(t time.Time) error  { return s.conn.SetReadDeadline(t) }
func (s *httpDirectStream) SetWriteDeadline(t time.Time) error { return s.conn.SetWriteDeadline(t) }

// BasicAuth 生成 Basic Auth 头
func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
