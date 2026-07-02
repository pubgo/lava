package tunnel_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pubgo/lava/v2/core/tunnel"
	_ "github.com/pubgo/lava/v2/core/tunnel/kcp"
	_ "github.com/pubgo/lava/v2/core/tunnel/quic"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelgateway"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux"
)

func TestTransport_Yamux(t *testing.T) { runTransportTest(t, "yamux", 20000) }
func TestTransport_QUIC(t *testing.T)  { runTransportTest(t, "quic", 20100) }
func TestTransport_KCP(t *testing.T)   { runTransportTest(t, "kcp", 20300) }

func runTransportTest(t *testing.T, transport string, basePort int) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// QUIC 使用自签证书，客户端需要跳过验证
	var transOpts *tunnel.TransportOptions
	if transport == tunnel.TransportQUIC {
		transOpts = &tunnel.TransportOptions{Insecure: true}
	}

	backendAddr := fmt.Sprintf("127.0.0.1:%d", basePort+81)
	gatewayAddr := fmt.Sprintf("127.0.0.1:%d", basePort)
	proxyAddr := fmt.Sprintf("http://127.0.0.1:%d", basePort+80)

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(map[string]string{"transport": transport}); err != nil {
			t.Logf("encode response failed: %v", err)
		}
	})
	backendSrv := &http.Server{Addr: backendAddr, Handler: mux}
	go func() {
		if err := backendSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("backend server error: %v", err)
		}
	}()
	t.Cleanup(func() {
		if err := backendSrv.Shutdown(ctx); err != nil {
			t.Logf("backend shutdown failed: %v", err)
		}
	})
	time.Sleep(100 * time.Millisecond)

	gw := tunnelgateway.New(&tunnelgateway.Config{
		ListenAddr:       gatewayAddr,
		Transport:        transport,
		TransportOptions: transOpts,
		HTTPPort:         basePort + 80,
	})
	if err := gw.Start(ctx); err != nil {
		t.Fatalf("[%s] Gateway: %v", transport, err)
	}
	t.Cleanup(func() {
		if err := gw.Stop(ctx); err != nil {
			t.Logf("[%s] Gateway stop failed: %v", transport, err)
		}
	})

	agent := tunnelagent.New(&tunnelagent.Config{
		GatewayAddr:      gatewayAddr,
		Transport:        transport,
		TransportOptions: transOpts,
		ServiceName:      transport + "-svc",
		Endpoints:        []tunnel.EndpointConfig{{Type: "http", LocalAddr: backendAddr}},
	})
	if err := agent.Start(ctx); err != nil {
		t.Fatalf("[%s] Agent: %v", transport, err)
	}
	t.Cleanup(func() {
		if err := agent.Stop(ctx); err != nil {
			t.Logf("[%s] Agent stop failed: %v", transport, err)
		}
	})
	time.Sleep(500 * time.Millisecond)

	if len(gw.Services()) == 0 {
		t.Fatalf("[%s] No services", transport)
	}

	resp, err := http.Get(proxyAddr + "/" + transport + "-svc/hello")
	if err != nil {
		t.Fatalf("[%s] Request: %v", transport, err)
	}
	t.Cleanup(func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("[%s] response close failed: %v", transport, err)
		}
	})
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("[%s] read response failed: %v", transport, err)
	}
	t.Logf("[%s] %s", transport, body)

	if resp.StatusCode != 200 {
		t.Errorf("[%s] Status %d", transport, resp.StatusCode)
	}
}

func TestAgentProxy_MultiBackends(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for i := 1; i <= 3; i++ {
		port := 21080 + i
		name := fmt.Sprintf("b%d", i)
		mux := http.NewServeMux()
		mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewEncoder(w).Encode(map[string]string{"backend": name}); err != nil {
				t.Logf("encode response failed: %v", err)
			}
		})
		srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port), Handler: mux}
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				t.Logf("backend server error: %v", err)
			}
		}()
		t.Cleanup(func() {
			if err := srv.Shutdown(ctx); err != nil {
				t.Logf("backend shutdown failed: %v", err)
			}
		})
	}
	time.Sleep(100 * time.Millisecond)

	gw := tunnelgateway.New(&tunnelgateway.Config{
		ListenAddr: "127.0.0.1:21000",
		Transport:  "yamux",
		HTTPPort:   21080,
	})
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := gw.Stop(ctx); err != nil {
			t.Logf("gateway stop failed: %v", err)
		}
	})

	var agents []*tunnelagent.Agent
	for i := 1; i <= 3; i++ {
		agent := tunnelagent.New(&tunnelagent.Config{
			GatewayAddr: "127.0.0.1:21000",
			Transport:   "yamux",
			ServiceName: fmt.Sprintf("svc-%d", i),
			Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: fmt.Sprintf("127.0.0.1:%d", 21080+i)}},
		})
		if err := agent.Start(ctx); err != nil {
			t.Fatal(err)
		}
		agents = append(agents, agent)
	}
	defer func() {
		for _, a := range agents {
			if err := a.Stop(ctx); err != nil {
				t.Logf("agent stop failed: %v", err)
			}
		}
	}()
	time.Sleep(500 * time.Millisecond)

	for i := 1; i <= 3; i++ {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:21080/svc-%d/api", i))
		if err != nil {
			t.Errorf("svc-%d: %v", i, err)
			continue
		}
		if err := resp.Body.Close(); err != nil {
			t.Errorf("svc-%d: close failed: %v", i, err)
		}
		if resp.StatusCode != 200 {
			t.Errorf("svc-%d: status %d", i, resp.StatusCode)
		}
	}
}

func TestAgentProxy_POST(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Logf("read request failed: %v", err)
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]any{"method": r.Method, "body": string(body)}); err != nil {
			t.Logf("encode response failed: %v", err)
		}
	})
	srv := &http.Server{Addr: "127.0.0.1:22081", Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("backend server error: %v", err)
		}
	}()
	t.Cleanup(func() {
		if err := srv.Shutdown(ctx); err != nil {
			t.Logf("backend shutdown failed: %v", err)
		}
	})
	time.Sleep(100 * time.Millisecond)

	gw := tunnelgateway.New(&tunnelgateway.Config{ListenAddr: "127.0.0.1:22000", Transport: "yamux", HTTPPort: 22080})
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := gw.Stop(ctx); err != nil {
			t.Logf("gateway stop failed: %v", err)
		}
	})

	agent := tunnelagent.New(&tunnelagent.Config{
		GatewayAddr: "127.0.0.1:22000",
		Transport:   "yamux",
		ServiceName: "post-svc",
		Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: "127.0.0.1:22081"}},
	})
	if err := agent.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := agent.Stop(ctx); err != nil {
			t.Logf("agent stop failed: %v", err)
		}
	})
	time.Sleep(500 * time.Millisecond)

	resp, err := http.Post("http://127.0.0.1:22080/post-svc/echo", "application/json", strings.NewReader(`{"test":"data"}`))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("response close failed: %v", err)
		}
	})

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if result["method"] != "POST" {
		t.Errorf("method=%v", result["method"])
	}
}

func TestAgentProxy_LargeBody(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Logf("read request failed: %v", err)
			return
		}
		if _, err := w.Write(body); err != nil {
			t.Logf("write response failed: %v", err)
		}
	})
	srv := &http.Server{Addr: "127.0.0.1:23081", Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("backend server error: %v", err)
		}
	}()
	t.Cleanup(func() {
		if err := srv.Shutdown(ctx); err != nil {
			t.Logf("backend shutdown failed: %v", err)
		}
	})
	time.Sleep(100 * time.Millisecond)

	gw := tunnelgateway.New(&tunnelgateway.Config{ListenAddr: "127.0.0.1:23000", Transport: "yamux", HTTPPort: 23080})
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := gw.Stop(ctx); err != nil {
			t.Logf("gateway stop failed: %v", err)
		}
	})

	agent := tunnelagent.New(&tunnelagent.Config{
		GatewayAddr: "127.0.0.1:23000",
		Transport:   "yamux",
		ServiceName: "upload-svc",
		Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: "127.0.0.1:23081"}},
	})
	if err := agent.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := agent.Stop(ctx); err != nil {
			t.Logf("agent stop failed: %v", err)
		}
	})
	time.Sleep(500 * time.Millisecond)

	for _, size := range []int{1024, 10 * 1024, 100 * 1024} {
		t.Run(fmt.Sprintf("%dKB", size/1024), func(t *testing.T) {
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}
			resp, err := http.Post("http://127.0.0.1:23080/upload-svc/upload", "application/octet-stream", bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("response close failed: %v", err)
				}
			})
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read response failed: %v", err)
			}
			if !bytes.Equal(data, body) {
				t.Error("body mismatch")
			}
		})
	}
}

func TestAgentProxy_Concurrent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var count int32
	var mu sync.Mutex
	mux := http.NewServeMux()
	mux.HandleFunc("/count", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		count++
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		if _, err := w.Write([]byte("ok")); err != nil {
			t.Logf("write response failed: %v", err)
		}
	})
	srv := &http.Server{Addr: "127.0.0.1:24081", Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("backend server error: %v", err)
		}
	}()
	t.Cleanup(func() {
		if err := srv.Shutdown(ctx); err != nil {
			t.Logf("backend shutdown failed: %v", err)
		}
	})
	time.Sleep(100 * time.Millisecond)

	gw := tunnelgateway.New(&tunnelgateway.Config{ListenAddr: "127.0.0.1:24000", Transport: "yamux", HTTPPort: 24080})
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := gw.Stop(ctx); err != nil {
			t.Logf("gateway stop failed: %v", err)
		}
	})

	agent := tunnelagent.New(&tunnelagent.Config{
		GatewayAddr: "127.0.0.1:24000",
		Transport:   "yamux",
		ServiceName: "concurrent-svc",
		Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: "127.0.0.1:24081"}},
	})
	if err := agent.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := agent.Stop(ctx); err != nil {
			t.Logf("agent stop failed: %v", err)
		}
	})
	time.Sleep(500 * time.Millisecond)

	n := 50
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get("http://127.0.0.1:24080/concurrent-svc/count")
			if err != nil {
				errs <- err
				return
			}
			if err := resp.Body.Close(); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
	mu.Lock()
	t.Logf("Total: %d", count)
	mu.Unlock()
}

func TestAgentProxy_TCP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ln, err := net.Listen("tcp", "127.0.0.1:26081")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := ln.Close(); err != nil {
			t.Logf("listener close failed: %v", err)
		}
	})
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer func() {
					if err := c.Close(); err != nil {
						t.Logf("conn close failed: %v", err)
					}
				}()
				if _, err := io.Copy(c, c); err != nil {
					t.Logf("echo copy failed: %v", err)
				}
			}(conn)
		}
	}()

	gw := tunnelgateway.New(&tunnelgateway.Config{ListenAddr: "127.0.0.1:26000", Transport: "yamux", HTTPPort: 26080})
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := gw.Stop(ctx); err != nil {
			t.Logf("gateway stop failed: %v", err)
		}
	})

	agent := tunnelagent.New(&tunnelagent.Config{
		GatewayAddr: "127.0.0.1:26000",
		Transport:   "yamux",
		ServiceName: "tcp-svc",
		Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: "127.0.0.1:26081"}},
	})
	if err := agent.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := agent.Stop(ctx); err != nil {
			t.Logf("agent stop failed: %v", err)
		}
	})
	time.Sleep(500 * time.Millisecond)

	if len(gw.Services()) == 0 {
		t.Fatal("no services")
	}
	t.Logf("services: %d", len(gw.Services()))
}

func TestMultipleAgents(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	gw := tunnelgateway.New(&tunnelgateway.Config{ListenAddr: "127.0.0.1:27000", Transport: "yamux", HTTPPort: 27080})
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := gw.Stop(ctx); err != nil {
			t.Logf("gateway stop failed: %v", err)
		}
	})

	n := 5
	var agents []*tunnelagent.Agent
	var servers []*http.Server

	for i := 0; i < n; i++ {
		port := 27081 + i
		name := fmt.Sprintf("multi-svc-%d", i)
		mux := http.NewServeMux()
		mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewEncoder(w).Encode(map[string]string{"svc": name}); err != nil {
				t.Logf("encode response failed: %v", err)
			}
		})
		srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port), Handler: mux}
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				t.Logf("backend server error: %v", err)
			}
		}()
		servers = append(servers, srv)

		agent := tunnelagent.New(&tunnelagent.Config{
			GatewayAddr: "127.0.0.1:27000",
			Transport:   "yamux",
			ServiceName: name,
			Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: fmt.Sprintf("127.0.0.1:%d", port)}},
		})
		if err := agent.Start(ctx); err != nil {
			t.Fatal(err)
		}
		agents = append(agents, agent)
	}
	defer func() {
		for _, a := range agents {
			if err := a.Stop(ctx); err != nil {
				t.Logf("agent stop failed: %v", err)
			}
		}
		for _, s := range servers {
			if err := s.Shutdown(ctx); err != nil {
				t.Logf("server shutdown failed: %v", err)
			}
		}
	}()
	time.Sleep(1 * time.Second)

	if len(gw.Services()) != n {
		t.Errorf("expected %d services, got %d", n, len(gw.Services()))
	}
	t.Logf("services: %d", len(gw.Services()))

	for i := 0; i < n; i++ {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:27080/multi-svc-%d/info", i))
		if err != nil {
			t.Errorf("svc %d: %v", i, err)
			continue
		}
		if resp != nil {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("svc %d: close failed: %v", i, err)
			}
			if resp.StatusCode != 200 {
				t.Errorf("svc %d: %d", i, resp.StatusCode)
			}
		}
	}
}

func TestProxyAuth_HTTP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	backendAddr := "127.0.0.1:27481"
	gatewayAddr := "127.0.0.1:27480"
	proxyPort := 27482
	token := "proxy-secret"

	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("authed"))
	})
	backend := &http.Server{Addr: backendAddr, Handler: mux}
	go func() { _ = backend.ListenAndServe() }()
	defer func() { _ = backend.Close() }()
	time.Sleep(100 * time.Millisecond)

	gw := tunnelgateway.NewGateway(&tunnel.GatewayConfig{
		ListenAddr: gatewayAddr,
		Transport:  tunnel.TransportYamux,
		HTTPPort:   proxyPort,
	})
	if err := tunnel.ConfigureGatewayAuth(gw, token); err != nil {
		t.Fatal(err)
	}
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = gw.Stop(context.Background()) }()

	agent, err := tunnelagent.Standalone(ctx, tunnelagent.StandaloneOptions{
		GatewayAddr: gatewayAddr,
		ServiceName: "auth-svc",
		AuthToken:   token,
		Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: backendAddr}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = agent.Stop(context.Background()) }()
	time.Sleep(500 * time.Millisecond)

	proxyBase := fmt.Sprintf("http://127.0.0.1:%d", proxyPort)
	resp, err := http.Get(proxyBase + "/auth-svc/ok")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no token: status=%d want 401", resp.StatusCode)
	}

	resp, err = tunnel.GetService(nil, proxyBase, "auth-svc", "/ok", token)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || string(body) != "authed" {
		t.Fatalf("status=%d body=%q", resp.StatusCode, body)
	}
}

func TestGRPCProxy_HealthCheck(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	token := "grpc-secret"
	grpcBackend := "127.0.0.1:27591"
	gatewayTunnel := "127.0.0.1:27590"
	grpcProxyPort := 27592

	srv, stopBackend := startGRPCHealthServer(t, grpcBackend)
	defer stopBackend()
	defer srv.Stop()

	gw := tunnelgateway.NewGateway(&tunnel.GatewayConfig{
		ListenAddr: gatewayTunnel,
		Transport:  tunnel.TransportYamux,
		GRPCPort:   grpcProxyPort,
	})
	if err := tunnel.ConfigureGatewayAuth(gw, token); err != nil {
		t.Fatal(err)
	}
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = gw.Stop(context.Background()) }()

	agent, err := tunnelagent.Standalone(ctx, tunnelagent.StandaloneOptions{
		GatewayAddr: gatewayTunnel,
		ServiceName: "grpc-health",
		AuthToken:   token,
		Endpoints:   []tunnel.EndpointConfig{{Type: "grpc", LocalAddr: grpcBackend}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = agent.Stop(context.Background()) }()
	time.Sleep(500 * time.Millisecond)

	cc, client, err := grpcHealthDial(ctx, fmt.Sprintf("127.0.0.1:%d", grpcProxyPort), "grpc-health", token)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cc.Close() }()

	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("status=%v", resp.GetStatus())
	}
}
