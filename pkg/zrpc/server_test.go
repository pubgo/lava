package zrpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/pubgo/lava/v2/pkg/zrpc"
)

func TestRegisterUnary(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("nats not available:", err)
	}
	defer nc.Close()

	srv := zrpc.NewServer(nc)
	defer srv.Close()

	subject := "svc.test.Runtime/Echo"
	queue := "test.runtime"
	if err := zrpc.RegisterUnary(srv, subject, queue,
		func() *wrapperspb.StringValue { return &wrapperspb.StringValue{} },
		func(_ context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
			return &wrapperspb.StringValue{Value: "echo:" + req.Value}, nil
		},
	); err != nil {
		t.Fatal(err)
	}

	data, err := proto.Marshal(&wrapperspb.StringValue{Value: "hi"})
	if err != nil {
		t.Fatal(err)
	}

	msg, err := nc.Request(subject, data, nats.DefaultTimeout)
	if err != nil {
		t.Fatal(err)
	}

	var resp wrapperspb.StringValue
	if err := proto.Unmarshal(msg.Data, &resp); err != nil {
		t.Fatal(err)
	}

	if resp.Value != "echo:hi" {
		t.Fatalf("unexpected response: %q", resp.Value)
	}
}

func TestHandleUnaryStatusError(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("nats not available:", err)
	}
	defer nc.Close()

	srv := zrpc.NewServer(nc)
	defer srv.Close()

	subject := "svc.test.Runtime/Bad"
	queue := "test.runtime"
	if err := zrpc.RegisterUnary(srv, subject, queue,
		func() *wrapperspb.StringValue { return &wrapperspb.StringValue{} },
		func(context.Context, *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
			return nil, zrpc.Errorf(zrpc.CodeInvalidArgument, "nope")
		},
	); err != nil {
		t.Fatal(err)
	}

	data, err := proto.Marshal(&wrapperspb.StringValue{Value: "x"})
	if err != nil {
		t.Fatal(err)
	}

	msg, err := nc.Request(subject, data, nats.DefaultTimeout)
	if err != nil {
		t.Fatal(err)
	}

	if got := msg.Header.Get(zrpc.HeaderStatusCode); got != "3" {
		t.Fatalf("unexpected code: %s", got)
	}
}

func TestClientCallUnary(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("nats not available:", err)
	}
	defer nc.Close()

	srv := zrpc.NewServer(nc)
	defer srv.Close()

	subject := "svc.test.Runtime/Client"
	queue := "test.runtime"
	if err := zrpc.RegisterUnary(srv, subject, queue,
		func() *wrapperspb.StringValue { return &wrapperspb.StringValue{} },
		func(_ context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
			return &wrapperspb.StringValue{Value: "resp:" + req.Value}, nil
		},
	); err != nil {
		t.Fatal(err)
	}

	cli := zrpc.NewClient(nc)
	var resp wrapperspb.StringValue
	if err := cli.CallUnary(context.Background(), subject, time.Second, &wrapperspb.StringValue{Value: "ok"}, &resp); err != nil {
		t.Fatal(err)
	}

	if resp.Value != "resp:ok" {
		t.Fatalf("unexpected response: %q", resp.Value)
	}
}
