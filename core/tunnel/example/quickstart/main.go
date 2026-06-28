// Demo: 最简上手 —— Standalone Agent + HTTP 代理（单进程跑通）。
//
// 演示：
//   - tunnelgateway 启动网关
//   - tunnelagent.Standalone 一行启动 Agent 并注册服务
//   - tunnel.GetService 经网关 HTTP 代理访问后端
//
// 运行：
//
//	go run ./core/tunnel/example/quickstart
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
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux" // 注册 yamux 传输
)

const (
	gatewayAddr = "127.0.0.1:17100" // Agent 连接端口
	httpPort    = 18100             // 对外 HTTP 代理端口
	backendAddr = "127.0.0.1:18101" // 本地后端服务
	serviceName = "hello-svc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1) 本地后端服务（被 tunnel 代理的目标）
	startBackend(ctx)

	// 2) 网关
	gw := tunnelgateway.NewGateway(&tunnel.GatewayConfig{
		ListenAddr: gatewayAddr,
		Transport:  tunnel.TransportYamux,
		HTTPPort:   httpPort,
	})
	if err := gw.Start(ctx); err != nil {
		log.Fatalf("gateway: %v", err)
	}
	defer gw.Stop(context.Background())

	// 3) Agent：一行启动 + 注册服务
	agent, err := tunnelagent.Standalone(ctx, tunnelagent.StandaloneOptions{
		GatewayAddr: gatewayAddr,
		ServiceName: serviceName,
		Endpoints: []tunnel.EndpointConfig{
			{Type: "http", LocalAddr: backendAddr},
		},
	})
	if err != nil {
		log.Fatalf("agent: %v", err)
	}
	defer agent.Stop(context.Background())

	time.Sleep(500 * time.Millisecond) // 等注册完成

	// 4) 经网关访问后端
	proxyBase := fmt.Sprintf("http://127.0.0.1:%d", httpPort)
	resp, err := tunnel.GetService(nil, proxyBase, serviceName, "/hello", "")
	if err != nil {
		log.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Println("----------------------------------------")
	fmt.Printf("GET %s\n", tunnel.ServiceURL(proxyBase, serviceName, "/hello"))
	fmt.Printf("status: %d\n", resp.StatusCode)
	fmt.Printf("body:   %s\n", body)
	fmt.Println("----------------------------------------")
	fmt.Printf("手动验证: curl %s/%s/hello\n", proxyBase, serviceName)
}

func startBackend(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "hello from backend")
	})
	srv := &http.Server{Addr: backendAddr, Handler: mux}
	go func() { _ = srv.ListenAndServe() }()
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	time.Sleep(100 * time.Millisecond)
}
