// Demo: TLS 传输 —— Agent ↔ Gateway 之间启用 TLS（yamux over TLS）。
//
// 演示：
//   - 运行时生成自签证书（仅 demo；生产用真实证书）
//   - GatewayConfig/AgentConfig 的 TLS 字段（含 MinVersion）
//   - TLS 加密的隧道上正常代理 HTTP
//
// 运行：
//
//	go run ./core/tunnel/example/tls
package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelgateway"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux"
)

const (
	gatewayAddr = "127.0.0.1:17400"
	httpPort    = 18400
	backendAddr = "127.0.0.1:18401"
	serviceName = "tls-svc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	certFile, keyFile, cleanup := mustGenerateCert()
	defer cleanup()

	startBackend(ctx)

	// 网关：服务端 TLS（提供证书）
	gw := tunnelgateway.NewGateway(&tunnel.GatewayConfig{
		ListenAddr: gatewayAddr,
		Transport:  tunnel.TransportYamux,
		HTTPPort:   httpPort,
		TLS: tunnel.TLSConfig{
			Enabled:    true,
			CertFile:   certFile,
			KeyFile:    keyFile,
			MinVersion: "TLS12",
		},
	})
	if err := gw.Start(ctx); err != nil {
		log.Fatalf("gateway: %v", err)
	}
	defer gw.Stop(context.Background())

	// Agent：客户端 TLS（demo 用 Insecure 跳过证书校验；生产配 CAFile）
	agent, err := tunnelagent.Standalone(ctx, tunnelagent.StandaloneOptions{
		GatewayAddr: gatewayAddr,
		ServiceName: serviceName,
		Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: backendAddr}},
		Config: &tunnel.AgentConfig{
			TLS: tunnel.TLSConfig{Enabled: true, Insecure: true, MinVersion: "TLS12"},
		},
	})
	if err != nil {
		log.Fatalf("agent: %v", err)
	}
	defer agent.Stop(context.Background())
	time.Sleep(500 * time.Millisecond)

	proxyBase := fmt.Sprintf("http://127.0.0.1:%d", httpPort)
	resp, err := tunnel.GetService(nil, proxyBase, serviceName, "/", "")
	if err != nil {
		log.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Println("----------------------------------------")
	fmt.Printf("TLS 隧道代理 OK, status=%d body=%q\n", resp.StatusCode, body)
	fmt.Println("Agent ↔ Gateway 链路已 TLS 加密（yamux over TLS）")
	fmt.Println("----------------------------------------")
}

func startBackend(ctx context.Context) {
	srv := &http.Server{Addr: backendAddr, Handler: http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "secure-backend") },
	)}
	go func() { _ = srv.ListenAndServe() }()
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	time.Sleep(100 * time.Millisecond)
}

// mustGenerateCert 生成自签 ECDSA 证书到临时文件（仅供 demo）。
func mustGenerateCert() (certFile, keyFile string, cleanup func()) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("gen key: %v", err)
	}
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "tunnel-demo"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:     []string{"localhost"},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		log.Fatalf("create cert: %v", err)
	}

	dir, err := os.MkdirTemp("", "tunnel-tls-demo")
	if err != nil {
		log.Fatalf("tmp dir: %v", err)
	}
	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")

	certOut, _ := os.Create(certFile)
	_ = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	_ = certOut.Close()

	keyBytes, _ := x509.MarshalECPrivateKey(key)
	keyOut, _ := os.Create(keyFile)
	_ = pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	_ = keyOut.Close()

	return certFile, keyFile, func() { _ = os.RemoveAll(dir) }
}
