package main

import (
	"context"
	"io"
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

	started := make(chan struct{})
	svc := zrpcs.New(zrpcs.Params{
		Log:     logger,
		Metric:  metric,
		Conf:    &zrpcs.Config{URL: "nats://127.0.0.1:4222"},
		Started: started,
		Registers: []zrpcs.RegisterFunc{
			func(srv *zrpc.Server) error {
				return zrpcdemov1.RegisterEchoServiceZrpcRoutes(srv, echoService{}, "")
			},
		},
	})

	errCh := make(chan error, 1)
	go func() { errCh <- svc.Serve(ctx) }()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for zrpc server to start")
	}
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

	ss, err := genCli.EchoStream(context.Background(), &zrpcdemov1.EchoRequest{Message: "HeLLo"})
	if err != nil {
		t.Fatal(err)
	}

	first, err := ss.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if first.GetMessage() != "HeLLo" {
		t.Fatalf("unexpected stream response 1: %q", first.GetMessage())
	}

	second, err := ss.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if second.GetMessage() != "HELLO" {
		t.Fatalf("unexpected stream response 2: %q", second.GetMessage())
	}

	third, err := ss.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if third.GetMessage() != "hello" {
		t.Fatalf("unexpected stream response 3: %q", third.GetMessage())
	}

	if _, err = ss.Recv(); err != io.EOF {
		t.Fatalf("expected EOF after server stream, got: %v", err)
	}

	cs, err := genCli.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if err = cs.Send(&zrpcdemov1.EchoRequest{Message: "a"}); err != nil {
		t.Fatal(err)
	}
	if err = cs.Send(&zrpcdemov1.EchoRequest{Message: "b"}); err != nil {
		t.Fatal(err)
	}
	collectResp, err := cs.CloseAndRecv()
	if err != nil {
		t.Fatal(err)
	}
	if collectResp.GetMessage() != "a,b" {
		t.Fatalf("unexpected collect response: %q", collectResp.GetMessage())
	}

	bs, err := genCli.Chat(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if err = bs.Send(&zrpcdemov1.EchoRequest{Message: "x"}); err != nil {
		t.Fatal(err)
	}
	b1, err := bs.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if b1.GetMessage() != "chat:x" {
		t.Fatalf("unexpected chat response 1: %q", b1.GetMessage())
	}

	if err = bs.Send(&zrpcdemov1.EchoRequest{Message: "y"}); err != nil {
		t.Fatal(err)
	}
	b2, err := bs.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if b2.GetMessage() != "chat:y" {
		t.Fatalf("unexpected chat response 2: %q", b2.GetMessage())
	}

	if err = bs.CloseSend(); err != nil {
		t.Fatal(err)
	}
	if _, err = bs.Recv(); err != io.EOF {
		t.Fatalf("expected EOF after bidi stream, got: %v", err)
	}
}
