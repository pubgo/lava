package main

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pubgo/funk/v2/log"
	"github.com/uber-go/tally/v4"

	"github.com/pubgo/lava/v2/clients/zrpcc"
	zrpcdemov1 "github.com/pubgo/lava/v2/pkg/proto/zrpcdemov1"
	"github.com/pubgo/lava/v2/pkg/zrpc"
	"github.com/pubgo/lava/v2/servers/zrpcs"
)

func TestGeneratedZrpcDemo(t *testing.T) {
	ncProbe, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("nats not available:", err)
	}
	ncProbe.Close()

	logger := log.GetLogger().WithName("zrpcdemo-test")
	metric := tally.NoopScope

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := zrpcs.New(zrpcs.Params{
		Log:    logger,
		Metric: metric,
		Conf:   &zrpcs.Config{URL: "nats://127.0.0.1:4222"},
		Registers: []zrpcs.RegisterFunc{
			func(srv *zrpc.Server) error {
				return zrpcdemov1.RegisterEchoServiceZrpcRoutes(srv, echoService{}, "")
			},
		},
	})

	errCh := make(chan error, 1)
	go func() { errCh <- svc.Serve(ctx) }()
	defer func() {
		cancel()
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("zrpc service failed: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for zrpc service shutdown")
		}
	}()

	cli := zrpcc.New(&zrpcc.Config{URL: "nats://127.0.0.1:4222", Timeout: time.Second}, zrpcc.Params{Log: logger, Metric: metric})
	defer func() { _ = cli.Close() }()

	if err := cli.Healthy(context.Background()); err != nil {
		t.Skip("nats not available:", err)
	}

	nc, err := cli.Conn()
	if err != nil {
		t.Fatal(err)
	}

	genCli := zrpcdemov1.NewEchoServiceZrpcClient(nc)

	echoResp, err := genCli.Echo(context.Background(), &zrpcdemov1.EchoRequest{Message: "hello"})
	if err != nil {
		t.Fatal(err)
	}

	if echoResp.GetMessage() != "hello" {
		t.Fatalf("unexpected echo response: %q", echoResp.GetMessage())
	}

	reverseResp, err := genCli.Reverse(context.Background(), &zrpcdemov1.EchoRequest{Message: "hello"})
	if err != nil {
		t.Fatal(err)
	}

	if reverseResp.GetMessage() != "olleh" {
		t.Fatalf("unexpected reverse response: %q", reverseResp.GetMessage())
	}
}
