package quicconn

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
	"time"

	"github.com/quic-go/quic-go"

	"github.com/pubgo/lava/v2/core/tunnel"
)

// Dial 在已有 UDP PacketConn 上向 remote 发起 QUIC 连接（ICE 打通后的承载层）。
func Dial(
	ctx context.Context,
	conn net.PacketConn,
	remote net.Addr,
	opts *tunnel.TransportOptions,
) (*quic.Conn, error) {
	tr := &quic.Transport{Conn: conn}
	tlsConf := clientTLS(opts)
	return tr.Dial(ctx, remote, tlsConf, quicConfig(opts))
}

// Listen 在已有 UDP PacketConn 上监听 QUIC 入站连接。
func Listen(
	conn net.PacketConn,
	opts *tunnel.TransportOptions,
) (*quic.Listener, error) {
	tr := &quic.Transport{Conn: conn}
	tlsConf, err := serverTLS(opts)
	if err != nil {
		return nil, err
	}
	return tr.Listen(tlsConf, quicConfig(opts))
}

func clientTLS(opts *tunnel.TransportOptions) *tls.Config {
	cfg := &tls.Config{NextProtos: []string{"lava-p2p"}}
	if opts != nil {
		cfg.InsecureSkipVerify = opts.Insecure
	}
	return cfg
}

// ServerTLS 构建服务端 TLS 配置。
func ServerTLS(opts *tunnel.TransportOptions) (*tls.Config, error) {
	if opts != nil && opts.CertFile != "" && opts.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile)
		if err != nil {
			return nil, err
		}
		return &tls.Config{Certificates: []tls.Certificate{cert}, NextProtos: []string{"lava-p2p"}}, nil
	}
	return devTLS()
}

func serverTLS(opts *tunnel.TransportOptions) (*tls.Config, error) {
	return ServerTLS(opts)
}

func devTLS() (*tls.Config, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"Lava P2P"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}
	cert := tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv}
	return &tls.Config{Certificates: []tls.Certificate{cert}, NextProtos: []string{"lava-p2p"}}, nil
}

// QUICConfig 构建 quic-go 配置（供测试与 transport 复用）。
func QUICConfig(opts *tunnel.TransportOptions) *quic.Config {
	return quicConfig(opts)
}

func quicConfig(opts *tunnel.TransportOptions) *quic.Config {
	cfg := &quic.Config{
		MaxIdleTimeout:  30 * time.Second,
		KeepAlivePeriod: 15 * time.Second,
	}
	if opts != nil && opts.MaxStreams > 0 {
		cfg.MaxIncomingStreams = int64(opts.MaxStreams)
	}
	return cfg
}
