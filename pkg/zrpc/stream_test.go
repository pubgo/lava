package zrpc_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/pubgo/lava/v2/pkg/zrpc"
)

func TestStreamBidi(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("nats not available:", err)
	}
	defer nc.Close()

	srv := zrpc.NewServer(nc)
	defer srv.Close()

	subject := "svc.test.Runtime/StreamBidi"
	queue := "test.runtime"
	if err = zrpc.RegisterStream(srv, subject, queue, func(ctx context.Context, stream *zrpc.ServerStream) error {
		for {
			var req wrapperspb.StringValue
			recvErr := stream.Recv(&req)
			if recvErr == io.EOF {
				return nil
			}
			if recvErr != nil {
				return recvErr
			}

			if sendErr := stream.Send(&wrapperspb.StringValue{Value: "echo:" + req.GetValue()}); sendErr != nil {
				return sendErr
			}
		}
	}); err != nil {
		t.Fatal(err)
	}

	cli := zrpc.NewClient(nc)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	st, err := cli.OpenStream(ctx, subject, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	if err = st.Send(&wrapperspb.StringValue{Value: "a"}); err != nil {
		t.Fatal(err)
	}
	if err = st.Send(&wrapperspb.StringValue{Value: "b"}); err != nil {
		t.Fatal(err)
	}
	if err = st.CloseSend(); err != nil {
		t.Fatal(err)
	}

	var r1 wrapperspb.StringValue
	if err = st.Recv(&r1); err != nil {
		t.Fatal(err)
	}
	if r1.GetValue() != "echo:a" {
		t.Fatalf("unexpected response: %q", r1.GetValue())
	}

	var r2 wrapperspb.StringValue
	if err = st.Recv(&r2); err != nil {
		t.Fatal(err)
	}
	if r2.GetValue() != "echo:b" {
		t.Fatalf("unexpected response: %q", r2.GetValue())
	}

	var end wrapperspb.StringValue
	if err = st.Recv(&end); err != io.EOF {
		t.Fatalf("expected EOF, got: %v", err)
	}
}

func TestStreamServerError(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("nats not available:", err)
	}
	defer nc.Close()

	srv := zrpc.NewServer(nc)
	defer srv.Close()

	subject := "svc.test.Runtime/StreamError"
	queue := "test.runtime"
	if err = zrpc.RegisterStream(srv, subject, queue, func(ctx context.Context, stream *zrpc.ServerStream) error {
		return zrpc.Errorf(zrpc.CodeInvalidArgument, "bad stream")
	}); err != nil {
		t.Fatal(err)
	}

	cli := zrpc.NewClient(nc)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	st, err := cli.OpenStream(ctx, subject, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	var rsp wrapperspb.StringValue
	err = st.Recv(&rsp)
	if err == nil {
		t.Fatal("expected stream error")
	}

	stErr, ok := err.(*zrpc.Status)
	if !ok {
		t.Fatalf("unexpected error type: %T, err=%v", err, err)
	}

	if stErr.Code != zrpc.CodeInvalidArgument {
		t.Fatalf("unexpected status code: %v", stErr.Code)
	}
}
