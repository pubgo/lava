package gateway

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestMux_GRPCPassthroughUnary(t *testing.T) {
	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find input type: %v", err)
	}
	outType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Struct")
	if err != nil {
		t.Fatalf("find output type: %v", err)
	}

	want := &structpb.Struct{Fields: map[string]*structpb.Value{"msg": structpb.NewStringValue("pong")}}

	mux := NewMux()
	mux.opts.handlers["/test.v1.Echo/Ping"] = &methodWrapper{
		srv: &serviceWrapper{
			opts:           mux.opts,
			remoteProxyCli: &fakeBackendConn{response: want},
		},
		grpcFullMethod: "/test.v1.Echo/Ping",
		inputType:      inType,
		outputType:     outType,
	}

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer(mux.GRPCServerOptions()...)
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer conn.Close()

	out := &structpb.Struct{}
	if err = conn.Invoke(ctx, "/test.v1.Echo/Ping", &emptypb.Empty{}, out); err != nil {
		t.Fatalf("invoke: %v", err)
	}
	if out.GetFields()["msg"].GetStringValue() != "pong" {
		t.Fatalf("unexpected response: %v", out)
	}
}

func TestMux_GRPCPassthroughUnknownMethod(t *testing.T) {
	mux := NewMux()
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer(mux.GRPCServerOptions()...)
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer conn.Close()

	err = conn.Invoke(ctx, "/unknown.Service/Method", &emptypb.Empty{}, &emptypb.Empty{})
	if err == nil {
		t.Fatal("expected error for unknown method")
	}
	if status.Code(err) != codes.Unimplemented {
		t.Fatalf("expected Unimplemented, got %v", err)
	}
}

type fakeBackendConn struct {
	response proto.Message
}

func (f *fakeBackendConn) Invoke(context.Context, string, any, any, ...grpc.CallOption) error {
	return status.Error(codes.Unimplemented, "use stream")
}

func (f *fakeBackendConn) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return &fakeBackendStream{response: f.response}, nil
}

type fakeBackendStream struct {
	response  proto.Message
	sentReply bool
}

func (f *fakeBackendStream) Header() (metadata.MD, error) { return nil, nil }
func (f *fakeBackendStream) Trailer() metadata.MD       { return nil }
func (f *fakeBackendStream) CloseSend() error           { return nil }
func (f *fakeBackendStream) Context() context.Context   { return context.Background() }
func (f *fakeBackendStream) SendMsg(any) error          { return nil }

func (f *fakeBackendStream) RecvMsg(m any) error {
	if f.sentReply {
		return io.EOF
	}
	f.sentReply = true
	pm, ok := m.(proto.Message)
	if !ok {
		return io.EOF
	}
	proto.Merge(pm, f.response)
	return nil
}
