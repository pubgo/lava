package zrpcc_test

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pubgo/funk/v2/log"
	"github.com/uber-go/tally/v4"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/pubgo/lava/v2/clients/zrpcc"
	zrpcdemov1 "github.com/pubgo/lava/v2/pkg/proto/zrpcdemov1"
	"github.com/pubgo/lava/v2/pkg/zrpc"
)

func TestClientClosedRejectsCalls(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("nats not available:", err)
	}
	defer nc.Close()

	srv := zrpc.NewServer(nc)
	defer srv.Close()

	subject := "svc.zrpcc.test/Closed"
	if err = zrpc.RegisterUnary(srv, subject, "zrpcc.test", func() *wrapperspb.StringValue {
		return &wrapperspb.StringValue{}
	}, func(context.Context, *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
		return &wrapperspb.StringValue{Value: "ok"}, nil
	}); err != nil {
		t.Fatal(err)
	}

	cli := zrpcc.New(&zrpcc.Config{URL: nats.DefaultURL, Timeout: time.Second}, zrpcc.Params{
		Log:    log.GetLogger(),
		Metric: tally.NoopScope,
	})

	if err = cli.Healthy(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err = cli.Close(); err != nil {
		t.Fatal(err)
	}

	var resp wrapperspb.StringValue
	if err = cli.CallUnary(context.Background(), subject, &wrapperspb.StringValue{Value: "x"}, &resp); err == nil {
		t.Fatal("expected error after client close")
	}
}

func TestGeneratedWithTimeoutReturnsCopy(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("nats not available:", err)
	}
	defer nc.Close()

	cli := zrpcdemov1.NewEchoServiceZrpcClient(nc)
	short := cli.WithTimeout(2 * time.Second)
	long := cli.WithTimeout(5 * time.Second)

	if short == cli || long == cli || short == long {
		t.Fatal("WithTimeout should return a new client instance")
	}
}
