package tunnelagent_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelgateway"
)

func TestStandaloneHTTPProxy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	backend := &http.Server{
		Addr: "127.0.0.1:27281",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, "standalone-ok")
		}),
	}
	go func() { _ = backend.ListenAndServe() }()
	defer func() { _ = backend.Close() }()
	time.Sleep(100 * time.Millisecond)

	gwAddr := "127.0.0.1:27280"
	gw := tunnelgateway.NewGateway(&tunnel.GatewayConfig{
		ListenAddr: gwAddr,
		Transport:  tunnel.TransportYamux,
		HTTPPort:   27282,
	})
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = gw.Stop(context.Background()) }()

	agent, err := tunnelagent.Standalone(ctx, tunnelagent.StandaloneOptions{
		GatewayAddr: gwAddr,
		ServiceName: "standalone-svc",
		Endpoints:   []tunnel.EndpointConfig{{Type: "http", LocalAddr: backend.Addr}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = agent.Stop(context.Background()) }()

	time.Sleep(500 * time.Millisecond)

	resp, err := tunnel.GetService(nil, "http://127.0.0.1:27282", "standalone-svc", "/", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "standalone-ok" {
		t.Fatalf("body=%q", body)
	}
}
