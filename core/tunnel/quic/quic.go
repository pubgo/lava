package quic

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func init() {
	tunnel.RegisterTransport(tunnel.TransportQUIC, NewTransport)
}

// NewTransport creates a new QUIC transport
func NewTransport(opts *tunnel.TransportOptions) (tunnel.Transport, error) {
	return &quicTransport{opts: opts}, nil
}

type quicTransport struct {
	opts *tunnel.TransportOptions
}

func (t *quicTransport) Name() string {
	return tunnel.TransportQUIC
}

func (t *quicTransport) Dial(ctx context.Context, addr string) (tunnel.Session, error) {
	tlsConfig := t.buildClientTLSConfig()

	quicConfig := t.buildQUICConfig()
	conn, err := quic.DialAddr(ctx, addr, tlsConfig, quicConfig)
	if err != nil {
		return nil, err
	}

	return &quicSession{conn: conn}, nil
}

func (t *quicTransport) Listen(ctx context.Context, addr string) (tunnel.Listener, error) {
	tlsConfig, err := t.buildServerTLSConfig()
	if err != nil {
		return nil, err
	}

	quicConfig := t.buildQUICConfig()
	ln, err := quic.ListenAddr(addr, tlsConfig, quicConfig)
	if err != nil {
		return nil, err
	}

	return &quicListener{listener: ln, ctx: ctx}, nil
}

func (t *quicTransport) buildClientTLSConfig() *tls.Config {
	tlsConfig := &tls.Config{
		NextProtos: []string{"lava-tunnel"},
	}
	if t.opts != nil {
		tlsConfig.InsecureSkipVerify = t.opts.Insecure
	}
	return tlsConfig
}

func (t *quicTransport) buildServerTLSConfig() (*tls.Config, error) {
	if t.opts == nil || t.opts.CertFile == "" || t.opts.KeyFile == "" {
		// 使用自签名证书进行开发/测试
		return generateSelfSignedTLSConfig()
	}

	cert, err := tls.LoadX509KeyPair(t.opts.CertFile, t.opts.KeyFile)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"lava-tunnel"},
	}, nil
}

func (t *quicTransport) buildQUICConfig() *quic.Config {
	cfg := &quic.Config{
		MaxIdleTimeout:  30 * time.Second,
		KeepAlivePeriod: 15 * time.Second,
	}
	if t.opts != nil {
		if t.opts.MaxStreams > 0 {
			cfg.MaxIncomingStreams = int64(t.opts.MaxStreams)
			cfg.MaxIncomingUniStreams = int64(t.opts.MaxStreams)
		}
		if t.opts.KeepAliveInterval > 0 {
			cfg.KeepAlivePeriod = time.Duration(t.opts.KeepAliveInterval) * time.Second
		}
	}
	return cfg
}

// generateSelfSignedTLSConfig 生成自签名 TLS 配置（仅用于开发/测试）
func generateSelfSignedTLSConfig() (*tls.Config, error) {
	cert, err := generateSelfSignedCert()
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"lava-tunnel"},
	}, nil
}

// generateSelfSignedCert 生成自签名证书
func generateSelfSignedCert() (tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Lava Tunnel"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  priv,
	}, nil
}

type quicListener struct {
	listener *quic.Listener
	ctx      context.Context
}

func (l *quicListener) Accept() (tunnel.Session, error) {
	conn, err := l.listener.Accept(l.ctx)
	if err != nil {
		return nil, err
	}
	return &quicSession{conn: conn}, nil
}

func (l *quicListener) Close() error   { return l.listener.Close() }
func (l *quicListener) Addr() net.Addr { return l.listener.Addr() }

type quicSession struct {
	conn *quic.Conn
	mu   sync.Mutex
}

func (s *quicSession) Open(ctx context.Context) (tunnel.Stream, error) {
	stream, err := s.conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	return &quicStream{stream: stream, conn: s.conn}, nil
}

// OpenWithPriority 打开指定优先级的流（1-10，1最高）
func (s *quicSession) OpenWithPriority(ctx context.Context, priority int) (tunnel.Stream, error) {
	// QUIC 不支持优先级，直接调用 Open
	return s.Open(ctx)
}

func (s *quicSession) Accept() (tunnel.Stream, error) {
	stream, err := s.conn.AcceptStream(context.Background())
	if err != nil {
		return nil, err
	}
	return &quicStream{stream: stream, conn: s.conn}, nil
}

func (s *quicSession) Close() error {
	return s.conn.CloseWithError(0, "closed")
}

func (s *quicSession) IsClosed() bool {
	select {
	case <-s.conn.Context().Done():
		return true
	default:
		return false
	}
}

func (s *quicSession) NumStreams() int {
	// QUIC 不直接暴露活跃流数量，返回 -1 表示未知
	return -1
}

func (s *quicSession) LocalAddr() net.Addr  { return s.conn.LocalAddr() }
func (s *quicSession) RemoteAddr() net.Addr { return s.conn.RemoteAddr() }

type quicStream struct {
	stream *quic.Stream
	conn   *quic.Conn
}

func (s *quicStream) Read(p []byte) (int, error)  { return s.stream.Read(p) }
func (s *quicStream) Write(p []byte) (int, error) { return s.stream.Write(p) }
func (s *quicStream) Close() error                { return s.stream.Close() }
func (s *quicStream) LocalAddr() net.Addr         { return s.conn.LocalAddr() }
func (s *quicStream) RemoteAddr() net.Addr        { return s.conn.RemoteAddr() }

func (s *quicStream) SetDeadline(t time.Time) error {
	if err := s.stream.SetReadDeadline(t); err != nil {
		return err
	}
	return s.stream.SetWriteDeadline(t)
}

func (s *quicStream) SetReadDeadline(t time.Time) error  { return s.stream.SetReadDeadline(t) }
func (s *quicStream) SetWriteDeadline(t time.Time) error { return s.stream.SetWriteDeadline(t) }

// Priority 获取流优先级
func (s *quicStream) Priority() int {
	// QUIC 不支持优先级，返回默认值
	return 5
}
