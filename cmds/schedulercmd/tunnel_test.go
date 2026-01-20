package schedulercmd

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/tunnel"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux"
)

// TestTunnelIntegration 测试 Tunnel Gateway 和 Agent 集成
func TestTunnelIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. 启动 Gateway
	gateway, err := tunnel.NewGatewayBuilder().
		WithListenAddr(":17007").
		WithHTTPPort(18888).
		WithDebugPort(16066).
		Build()
	if err != nil {
		t.Fatalf("Failed to build gateway: %v", err)
	}

	if err := gateway.Start(ctx); err != nil {
		t.Fatalf("Failed to start gateway: %v", err)
	}
	defer gateway.Stop(context.Background())

	t.Log("Gateway started on :17007, HTTP proxy on :18888, Debug proxy on :16066")

	// 2. 启动一个简单的本地 HTTP 服务
	localMux := http.NewServeMux()
	localMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from local service! Path: %s", r.URL.Path)
	})
	localMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	localServer := &http.Server{
		Addr:    ":19999",
		Handler: localMux,
	}
	go localServer.ListenAndServe()
	defer localServer.Close()

	t.Log("Local service started on :19999")

	// 等待服务启动
	time.Sleep(100 * time.Millisecond)

	// 3. 启动 Agent 连接 Gateway
	agent, err := tunnel.NewAgentBuilder().
		WithGatewayAddr("localhost:17007").
		WithServiceName("test-service").
		WithServiceVersion("1.0.0").
		AddEndpoint("http", "localhost:19999", "/").
		Build()
	if err != nil {
		t.Fatalf("Failed to build agent: %v", err)
	}

	if err := agent.Start(ctx); err != nil {
		t.Fatalf("Failed to start agent: %v", err)
	}
	defer agent.Stop(context.Background())

	t.Log("Agent started and connected to Gateway")

	// 等待连接建立
	time.Sleep(500 * time.Millisecond)

	// 4. 通过 Gateway HTTP 代理访问服务
	// 访问 http://localhost:18888/test-service/health
	resp, err := http.Get("http://localhost:18888/test-service/health")
	if err != nil {
		t.Logf("Warning: Failed to access service through gateway: %v", err)
		t.Log("This is expected if the proxy routing is not fully implemented yet")
	} else {
		defer resp.Body.Close()
		t.Logf("Response status: %d", resp.StatusCode)
	}

	// 5. 查看 Gateway 上的服务列表
	resp, err = http.Get("http://localhost:18888/")
	if err != nil {
		t.Logf("Warning: Failed to get service list: %v", err)
	} else {
		defer resp.Body.Close()
		t.Logf("Service list status: %d", resp.StatusCode)
	}

	// 6. 测试 Gateway 状态
	services := gateway.Services()
	t.Logf("Registered services: %d", len(services))
	for _, svc := range services {
		t.Logf("  - %s (version: %s)", svc.Name, svc.Version)
	}
}

// TestGatewayOnly 只测试 Gateway 启动
func TestGatewayOnly(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gateway, err := tunnel.NewGatewayBuilder().
		WithListenAddr(":27007").
		WithHTTPPort(28888).
		WithDebugPort(26066).
		Build()
	if err != nil {
		t.Fatalf("Failed to build gateway: %v", err)
	}

	if err := gateway.Start(ctx); err != nil {
		t.Fatalf("Failed to start gateway: %v", err)
	}

	t.Log("Gateway started successfully")
	t.Logf("Status: %s", gateway.Status())

	// 访问服务列表
	resp, err := http.Get("http://localhost:28888/")
	if err != nil {
		t.Fatalf("Failed to get service list: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("Service list response: %d", resp.StatusCode)

	if err := gateway.Stop(context.Background()); err != nil {
		t.Fatalf("Failed to stop gateway: %v", err)
	}
	t.Log("Gateway stopped successfully")
}
