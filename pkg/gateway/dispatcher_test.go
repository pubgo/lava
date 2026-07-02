package gateway

import (
	"context"
	"io"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type fakeFrontendStream struct {
	ctx     context.Context
	recvEOF bool
	sent    []proto.Message
	header  metadata.MD
	trailer metadata.MD
}

func (f *fakeFrontendStream) SetHeader(md metadata.MD) error {
	f.header = metadata.Join(f.header, md)
	return nil
}

func (f *fakeFrontendStream) SendHeader(md metadata.MD) error {
	return f.SetHeader(md)
}

func (f *fakeFrontendStream) SetTrailer(md metadata.MD) {
	f.trailer = metadata.Join(f.trailer, md)
}

func (f *fakeFrontendStream) Context() context.Context {
	if f.ctx == nil {
		return context.Background()
	}
	return f.ctx
}

func (f *fakeFrontendStream) SendMsg(m any) error {
	if pm, ok := m.(proto.Message); ok {
		f.sent = append(f.sent, proto.Clone(pm))
	}
	return nil
}

func (f *fakeFrontendStream) RecvMsg(m any) error {
	if f.recvEOF {
		return io.EOF
	}
	f.recvEOF = true
	return nil
}

func TestDispatcher_DispatchBidi(t *testing.T) {
	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find input type: %v", err)
	}
	outType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Struct")
	if err != nil {
		t.Fatalf("find output type: %v", err)
	}

	fakeStream := &fakeClientStream{
		trailer: metadata.Pairs("grpc-status", "0"),
	}
	backend := &fakeClientConn{stream: fakeStream}
	frontend := &fakeFrontendStream{}
	op := &Operation{
		FullMethod: "/test.v1.StreamService/Bidi",
		InputType:  inType,
		OutputType: outType,
		StreamDesc: &grpc.StreamDesc{ServerStreams: true, ClientStreams: true},
	}

	d := NewDispatcher()
	if _, _, err = d.Dispatch(context.Background(), backend, frontend, op, nil); err != nil {
		t.Fatalf("Dispatch bidi failed: %v", err)
	}
}

func TestDispatcher_DispatchServerStream(t *testing.T) {
	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find input type: %v", err)
	}
	outType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Struct")
	if err != nil {
		t.Fatalf("find output type: %v", err)
	}

	fakeStream := &fakeClientStream{
		header: metadata.Pairs("x-stream", "header"),
		frames: []proto.Message{
			&structpb.Struct{Fields: map[string]*structpb.Value{"msg": structpb.NewStringValue("hello")}},
		},
	}
	backend := &fakeClientConn{stream: fakeStream}
	frontend := &fakeFrontendStream{}
	op := &Operation{
		FullMethod: "/test.v1.StreamService/Watch",
		InputType:  inType,
		OutputType: outType,
		StreamDesc: &grpc.StreamDesc{ServerStreams: true, ClientStreams: false},
	}

	d := NewDispatcher()
	if _, _, err = d.Dispatch(context.Background(), backend, frontend, op, &emptypb.Empty{}); err != nil {
		t.Fatalf("Dispatch server stream failed: %v", err)
	}
	if len(frontend.sent) != 1 {
		t.Fatalf("expected 1 response frame, got %d", len(frontend.sent))
	}
}
