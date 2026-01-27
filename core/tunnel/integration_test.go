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

	"github.com/pubgo/lava/v2/core/tunnel"
	_ "github.com/pubgo/lava/v2/core/tunnel/http"
	_ "github.com/pubgo/lava/v2/core/tunnel/kcp"
	_ "github.com/pubgo/lava/v2/core/tunnel/quic"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelgateway"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux"
)

func TestTransport_Yamux(t *testing.T) { runTransportTest(t, "yamux", 20000) }
func TestTransport_QUIC(t *testing.T)  { runTransportTest(t, "quic", 20100) }
func TestTransport_HTTP(t *testing.T)  { runTransportTest(t, "http", 20200) }
func TestTransport_KCP(t *testing.T)   { runTransportTest(t, "kcp", 20300) }

func runTransportTest(t *testing.T, transport string, basePort int) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	backendAddr := fmt.Sprintf("127.0.0.1:%d", basePort+81)
	gatewayAddr := fmt.Sprintf("127.0.0.1:%d", basePort)
	proxyAddr := fmt.Sprintf("http://127.0.0.1:%d", basePort+80)

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"transport": transport})
	})
	backendSrv := &http.Server{Addr: backendAddr, Handler: mux}
	go backendSrv.ListenAndServe()
	defer backendSrv.Shutdown(ctx)
	time.Sleep(100 * time.Millisecond)

	gw := tunnelgateway.New(&tunnelgateway.Config{
		ListenAddr: gatewayAddr,
		Transport:  transport,
		HTTPPort:   basePort + 80,
	})
	if err := gw.Start(ctx); err != nil {
		t.Fatalf("[%s] Gateway: %v", transport, err)
	}
	defer gw.Stop(ctx)

	agent := tunnelagent.New(&tunnelagent.Config{
		GatewayAddr: gatewayAddr,
		Transport:   transport,
		ServiceName: transport + "-svc",
		Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: backendAddr}},
	})
	if err := agent.Start(ctx); err != nil {
		t.Fatalf("[%s] Agent: %v", transport, err)
	}
	defer agent.Stop(ctx)
	time.Sleep(500 * time.Millisecond)

	if len(gw.Services()) == 0 {
		t.Fatalf("[%s] No services", transport)
	}

	resp, err := http.Get(proxyAddr + "/" + transport + "-svc/hello")
	if err != nil {
		t.Fatalf("[%s] Request: %v", transport, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
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
			json.NewEncoder(w).Encode(map[string]string{"backend": name})
		})
		srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port), Handler: mux}
		go srv.ListenAndServe()
		defer srv.Shutdown(ctx)
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
	defer gw.Stop(ctx)

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
			a.Stop(ctx)
		}
	}()
	time.Sleep(500 * time.Millisecond)

	for i := 1; i <= 3; i++ {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:21080/svc-%d/api", i))
		if err != nil {
			t.Errorf("svc-%d: %v", i, err)
			continue
		}
		resp.Body.Close()
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
		body, _ := io.ReadAll(r.Body)
		json.NewEncoder(w).Encode(map[string]any{"method": r.Method, "body": string(body)})
	})
	srv := &http.Server{Addr: "127.0.0.1:22081", Handler: mux}
	go srv.ListenAndServe()
	defer srv.Shutdown(ctx)
	time.Sleep(100 * time.Millisecond)

	gw := tunnelgateway.New(&tunnelgateway.Config{ListenAddr: "127.0.0.1:22000", Transport: "yamux", HTTPPort: 22080})
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer gw.Stop(ctx)

	agent := tunnelagent.New(&tunnelagent.Config{
		GatewayAddr: "127.0.0.1:22000",
		Transport:   "yamux",
		ServiceName: "post-svc",
		Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: "127.0.0.1:22081"}},
	})
	if err := agent.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer agent.Stop(ctx)
	time.Sleep(500 * time.Millisecond)

	resp, err := http.Post("http://127.0.0.1:22080/post-svc/echo", "application/json", strings.NewReader(`{"test":"data"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	if result["method"] != "POST" {
		t.Errorf("method=%v", result["method"])
	}
}

func TestAgentProxy_LargeBody(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
	})
	srv := &http.Server{Addr: "127.0.0.1:23081", Handler: mux}
	go srv.ListenAndServe()
	defer srv.Shutdown(ctx)
	time.Sleep(100 * time.Millisecond)

	gw := tunnelgateway.New(&tunnelgateway.Config{ListenAddr: "127.0.0.1:23000", Transport: "yamux", HTTPPort: 23080})
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer gw.Stop(ctx)

	agent := tunnelagent.New(&tunnelagent.Config{
		GatewayAddr: "127.0.0.1:23000",
		Transport:   "yamux",
		ServiceName: "upload-svc",
		Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: "127.0.0.1:23081"}},
	})
	if err := agent.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer agent.Stop(ctx)
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
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
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
		w.Write([]byte("ok"))
	})
	srv := &http.Server{Addr: "127.0.0.1:24081", Handler: mux}
	go srv.ListenAndServe()
	defer srv.Shutdown(ctx)
	time.Sleep(100 * time.Millisecond)

	gw := tunnelgateway.New(&tunnelgateway.Config{ListenAddr: "127.0.0.1:24000", Transport: "yamux", HTTPPort: 24080})
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer gw.Stop(ctx)

	agent := tunnelagent.New(&tunnelagent.Config{
		GatewayAddr: "127.0.0.1:24000",
		Transport:   "yamux",
		ServiceName: "concurrent-svc",
		Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: "127.0.0.1:24081"}},
	})
	if err := agent.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer agent.Stop(ctx)
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
			resp.Body.Close()
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
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()

	gw := tunnelgateway.New(&tunnelgateway.Config{ListenAddr: "127.0.0.1:26000", Transport: "yamux", HTTPPort: 26080})
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer gw.Stop(ctx)

	agent := tunnelagent.New(&tunnelagent.Config{
		GatewayAddr: "127.0.0.1:26000",
		Transport:   "yamux",
		ServiceName: "tcp-svc",
		Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: "127.0.0.1:26081"}},
	})
	if err := agent.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer agent.Stop(ctx)
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
	defer gw.Stop(ctx)

	n := 5
	var agents []*tunnelagent.Agent
	var servers []*http.Server

	for i := 0; i < n; i++ {
		port := 27081 + i
		name := fmt.Sprintf("multi-svc-%d", i)
		mux := http.NewServeMux()
		mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{"svc": name})
		})
		srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port), Handler: mux}
		go srv.ListenAndServe()
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
			a.Stop(ctx)
		}
		for _, s := range servers {
			s.Shutdown(ctx)
		}
	}()
	time.Sleep(1 * time.Second)

	if len(gw.Services()) != n {
		t.Errorf("expected %d services, got %d", n, len(gw.Services()))
	}
	t.Logf("services: %d", len(gw.Services()))

	for i := 0; i < n; i++ {
		resp, _ := http.Get(fmt.Sprintf("http://127.0.0.1:27080/multi-svc-%d/info", i))
		if resp != nil {
			resp.Body.Close()
			if resp.StatusCode != 200 {
				t.Errorf("svc %d: %d", i, resp.StatusCode)
			}
		}
	}
}
