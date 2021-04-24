package quick

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
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
	return quic.ListenAddr(addr, tlsConf, t.ToCfg())
}

// Listen creates a QUIC listener on the given network interface
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
	return quic.DialAddr(addr, tlsConf, t.ToCfg())
}

// DialConn creates a new QUIC connection
// it returns once the connection is established and secured with forward-secure keys
func (t Cfg) DialConn(addr string, tlsConf *tls.Config) (net.Conn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, err
	}

	// DialAddr returns once a forward-secure connection is established
	session, err := quic.Dial(udpConn, udpAddr, addr, tlsConf, t.ToCfg())
	if err != nil {
		return nil, err
	}

	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}

	return &conn{conn: udpConn, session: session, stream: stream}, nil
}

func (t Cfg) ToCfg() *quic.Config {
	var cfg = quic.Config{}

	xerror.Panic(merge.Copy(&cfg, &t))

	return &cfg
}

func GetDefaultCfg() Cfg {
	return Cfg{

	}
}

func generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	b := pem.Block{Type: "CERTIFICATE", Bytes: certDER}
	certPEM := pem.EncodeToMemory(&b)

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quick"},
	}, nil
}
