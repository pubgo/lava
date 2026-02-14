// Package main provides a simple example of using the tunnel package.
//
// This example demonstrates:
// - Starting a Gateway server
// - Connecting an Agent to the Gateway
// - Registering services and proxying HTTP requests
//
// Usage:
//
//	# Run all components together (gateway + agent + backend)
//	go run ./core/tunnel/example/main.go
//
//	# Run with different transport protocols
//	go run ./core/tunnel/example/main.go -transport=quic
//	go run ./core/tunnel/example/main.go -transport=kcp
//
//	# Run components separately (in different terminals)
//	go run ./core/tunnel/example/main.go -mode=backend
//	go run ./core/tunnel/example/main.go -mode=gateway
//	go run ./core/tunnel/example/main.go -mode=agent
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pubgo/lava/v2/core/tunnel"
	_ "github.com/pubgo/lava/v2/core/tunnel/kcp"
	_ "github.com/pubgo/lava/v2/core/tunnel/quic"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelgateway"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux"
)

var (
	mode        = flag.String("mode", "all", "Run mode: gateway, agent, backend, or all")
	transport   = flag.String("transport", "yamux", "Transport protocol: yamux, quic, kcp")
	gatewayAddr = flag.String("gateway-addr", "127.0.0.1:17000", "Gateway listen address")
	httpPort    = flag.Int("http-port", 18080, "HTTP proxy port")
	backendAddr = flag.String("backend-addr", "127.0.0.1:18081", "Backend service address")
	serviceName = flag.String("service", "demo-svc", "Service name for agent")
)

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	switch *mode {
	case "gateway":
		runGateway(ctx)
	case "agent":
		runAgent(ctx)
	case "backend":
		runBackend(ctx)
	case "all":
		runAll(ctx)
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}

func runGateway(ctx context.Context) {
	gw := tunnelgateway.New(&tunnelgateway.Config{
		ListenAddr: *gatewayAddr,
		Transport:  *transport,
		HTTPPort:   *httpPort,
	})

	if err := gw.Start(ctx); err != nil {
		log.Fatalf("Gateway start failed: %v", err)
	}
	defer func() {
		if err := gw.Stop(ctx); err != nil {
			log.Printf("Gateway stop failed: %v", err)
		}
	}()

	fmt.Printf("Gateway started on %s (transport: %s)\n", *gatewayAddr, *transport)
	fmt.Printf("HTTP proxy available at http://127.0.0.1:%d/<service-name>/<path>\n", *httpPort)

	<-ctx.Done()
}

func runAgent(ctx context.Context) {
	agent := tunnelagent.New(&tunnelagent.Config{
		GatewayAddr: *gatewayAddr,
		Transport:   *transport,
		ServiceName: *serviceName,
		Endpoints: []tunnel.EndpointConfig{
			{Type: "http", LocalAddr: *backendAddr},
		},
	})

	if err := agent.Start(ctx); err != nil {
		log.Fatalf("Agent start failed: %v", err)
	}
	defer func() {
		if err := agent.Stop(ctx); err != nil {
			log.Printf("Agent stop failed: %v", err)
		}
	}()

	fmt.Printf("Agent connected to %s (service: %s, backend: %s)\n", *gatewayAddr, *serviceName, *backendAddr)

	<-ctx.Done()
}

func runBackend(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(map[string]any{
			"path":    r.URL.Path,
			"method":  r.Method,
			"headers": r.Header,
			"time":    time.Now().Format(time.RFC3339),
		}); err != nil {
			log.Printf("encode response failed: %v", err)
		}
	})

	srv := &http.Server{Addr: *backendAddr, Handler: mux}
	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("backend shutdown failed: %v", err)
		}
	}()

	fmt.Printf("Backend server started on %s\n", *backendAddr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Backend server error: %v", err)
	}
}

func runAll(ctx context.Context) {
	fmt.Println("Starting backend server...")
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(map[string]any{
			"message": "Hello from backend!",
			"path":    r.URL.Path,
			"time":    time.Now().Format(time.RFC3339),
		}); err != nil {
			log.Printf("encode response failed: %v", err)
		}
	})
	backendSrv := &http.Server{Addr: *backendAddr, Handler: mux}
	go func() {
		if err := backendSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("backend server error: %v", err)
		}
	}()
	defer func() {
		if err := backendSrv.Shutdown(ctx); err != nil {
			log.Printf("backend shutdown failed: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	fmt.Println("Starting gateway...")
	gw := tunnelgateway.New(&tunnelgateway.Config{
		ListenAddr: *gatewayAddr,
		Transport:  *transport,
		HTTPPort:   *httpPort,
	})
	if err := gw.Start(ctx); err != nil {
		log.Fatalf("Gateway start failed: %v", err)
	}
	defer func() {
		if err := gw.Stop(ctx); err != nil {
			log.Printf("Gateway stop failed: %v", err)
		}
	}()

	fmt.Println("Starting agent...")
	agent := tunnelagent.New(&tunnelagent.Config{
		GatewayAddr: *gatewayAddr,
		Transport:   *transport,
		ServiceName: *serviceName,
		Endpoints: []tunnel.EndpointConfig{
			{Type: "http", LocalAddr: *backendAddr},
		},
	})
	if err := agent.Start(ctx); err != nil {
		log.Fatalf("Agent start failed: %v", err)
	}
	defer func() {
		if err := agent.Stop(ctx); err != nil {
			log.Printf("Agent stop failed: %v", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	fmt.Println()
	fmt.Println("========================================")
	fmt.Printf("Tunnel demo running (transport: %s)\n", *transport)
	fmt.Println("========================================")
	fmt.Printf("Gateway:     %s\n", *gatewayAddr)
	fmt.Printf("HTTP Proxy:  http://127.0.0.1:%d\n", *httpPort)
	fmt.Printf("Service:     %s\n", *serviceName)
	fmt.Printf("Backend:     %s\n", *backendAddr)
	fmt.Println()
	fmt.Printf("Try: curl http://127.0.0.1:%d/%s/hello\n", *httpPort, *serviceName)
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")

	<-ctx.Done()
}
