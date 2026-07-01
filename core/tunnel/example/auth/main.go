// Demo: 代理鉴权 —— 配置 TUNNEL_AUTH_TOKEN 后，HTTP 代理要求客户端携带 token。
//
// 演示：
//   - ConfigureGatewayAuth 启用网关 token 鉴权
//   - 无 token 请求被拒（401）
//   - 带 token 请求成功（Authorization: Bearer / X-Tunnel-Token）
//
// 运行：
//
//	go run ./core/tunnel/example/auth
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelgateway"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux"
)

const (
	gatewayAddr = "127.0.0.1:17200"
	httpPort    = 18200
	backendAddr = "127.0.0.1:18201"
	serviceName = "secure-svc"
	authToken   = "demo-secret-token"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startBackend(ctx)

	// 网关：启用 token 鉴权（注册 + 代理面都受保护）
	gw := tunnelgateway.NewGateway(&tunnel.GatewayConfig{
		ListenAddr: gatewayAddr,
		Transport:  tunnel.TransportYamux,
		HTTPPort:   httpPort,
	})
	if err := tunnel.ConfigureGatewayAuth(gw, authToken); err != nil {
		log.Fatalf("auth: %v", err)
	}
	if err := gw.Start(ctx); err != nil {
		log.Fatalf("gateway: %v", err)
	}
	defer func() { _ = gw.Stop(context.Background()) }()

	// Agent：注册时携带同一 token
	agent, err := tunnelagent.Standalone(ctx, tunnelagent.StandaloneOptions{
		GatewayAddr: gatewayAddr,
		ServiceName: serviceName,
		AuthToken:   authToken,
		Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: backendAddr}},
	})
	if err != nil {
		log.Fatalf("agent: %v", err)
	}
	defer func() { _ = agent.Stop(context.Background()) }()
	time.Sleep(500 * time.Millisecond)

	proxyBase := fmt.Sprintf("http://127.0.0.1:%d", httpPort)

	// 1) 无 token → 401
	noTokenResp, err := http.Get(proxyBase + "/" + serviceName + "/ping")
	if err != nil {
		log.Fatalf("no-token request: %v", err)
	}
	_ = noTokenResp.Body.Close()
	fmt.Printf("[无 token]  status=%d (期望 401)\n", noTokenResp.StatusCode)

	// 2) 带 token → 200
	okResp, err := tunnel.GetService(nil, proxyBase, serviceName, "/ping", authToken)
	if err != nil {
		log.Fatalf("token request: %v", err)
	}
	defer func() { _ = okResp.Body.Close() }()
	body, _ := io.ReadAll(okResp.Body)
	fmt.Printf("[带 token]  status=%d body=%q (期望 200)\n", okResp.StatusCode, body)

	fmt.Println("----------------------------------------")
	fmt.Printf("curl -H 'Authorization: Bearer %s' %s/%s/ping\n", authToken, proxyBase, serviceName)
}

func startBackend(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "pong")
	})
	srv := &http.Server{Addr: backendAddr, Handler: mux}
	go func() { _ = srv.ListenAndServe() }()
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	time.Sleep(100 * time.Millisecond)
}
