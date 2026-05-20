package gateway

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type fakeClientConn struct {
	stream grpc.ClientStream
}

func (f *fakeClientConn) Invoke(context.Context, string, any, any, ...grpc.CallOption) error {
	return nil
}

func (f *fakeClientConn) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return f.stream, nil
}

type fakeClientStream struct {
	header           metadata.MD
	trailer          metadata.MD
	frames           []proto.Message
	readIdx          int
	closed           bool
	sentReqs         []proto.Message
	headerCalled     bool
	earlyHeaderFetch bool
	failOnEarlyHead  bool
}

func (f *fakeClientStream) Header() (metadata.MD, error) {
	f.headerCalled = true
	if f.readIdx == 0 {
		f.earlyHeaderFetch = true
	}
	return f.header, nil
}

func (f *fakeClientStream) Trailer() metadata.MD {
	return f.trailer
}

func (f *fakeClientStream) CloseSend() error {
	f.closed = true
	return nil
}

func (f *fakeClientStream) Context() context.Context {
	return context.Background()
}

func (f *fakeClientStream) SendMsg(m any) error {
	if pm, ok := m.(proto.Message); ok {
		f.sentReqs = append(f.sentReqs, proto.Clone(pm))
	}
	return nil
}

func (f *fakeClientStream) RecvMsg(m any) error {
	if f.failOnEarlyHead && f.earlyHeaderFetch {
		return io.ErrUnexpectedEOF
	}

	if f.readIdx >= len(f.frames) {
		return io.EOF
	}
	pm, ok := m.(proto.Message)
	if !ok {
		return io.EOF
	}
	frame := f.frames[f.readIdx]
	f.readIdx++
	b, err := proto.Marshal(frame)
	if err != nil {
		return err
	}
	return proto.Unmarshal(b, pm)
}

func TestInvokeResponseStream_DoesNotPrefetchHeaderBeforeFirstFrame(t *testing.T) {
	mux := NewMux()

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
		failOnEarlyHead: true,
	}

	method := &methodWrapper{
		srv:            &serviceWrapper{opts: mux.opts, remoteProxyCli: &fakeClientConn{stream: fakeStream}},
		grpcStreamDesc: &grpc.StreamDesc{ServerStreams: true, ClientStreams: false},
		grpcFullMethod: "/test.v1.StreamService/Watch",
		inputType:      inType,
		outputType:     outType,
	}
	mux.opts.handlers[method.grpcFullMethod] = method

	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")

	stream := &streamHTTP{handler: ctx, ctx: context.Background(), method: method}

	if err = mux.invokeResponseStream(stream, &emptypb.Empty{}); err != nil {
		t.Fatalf("invokeResponseStream failed: %v", err)
	}

	if !fakeStream.headerCalled {
		t.Fatal("expected header to be fetched eventually")
	}
	if fakeStream.earlyHeaderFetch {
		t.Fatal("header was prefetched before first frame")
	}

	body := string(ctx.Response().Body())
	if !strings.Contains(body, "\"msg\":\"hello\"") {
		t.Fatalf("unexpected response body: %q", body)
	}
}

func TestInvokeResponseStream_AllowsPreSentHeaderAndStreamsJSON(t *testing.T) {
	mux := NewMux()

	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find input type: %v", err)
	}
	outType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Struct")
	if err != nil {
		t.Fatalf("find output type: %v", err)
	}

	fakeStream := &fakeClientStream{
		header:  metadata.Pairs("x-stream", "header"),
		trailer: metadata.Pairs("grpc-status", "0"),
		frames: []proto.Message{
			&structpb.Struct{Fields: map[string]*structpb.Value{"msg": structpb.NewStringValue("a")}},
			&structpb.Struct{Fields: map[string]*structpb.Value{"msg": structpb.NewStringValue("b")}},
		},
	}

	method := &methodWrapper{
		srv:            &serviceWrapper{opts: mux.opts, remoteProxyCli: &fakeClientConn{stream: fakeStream}},
		grpcStreamDesc: &grpc.StreamDesc{ServerStreams: true, ClientStreams: false},
		grpcFullMethod: "/test.v1.StreamService/Watch",
		inputType:      inType,
		outputType:     outType,
	}
	mux.opts.handlers[method.grpcFullMethod] = method

	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")

	stream := &streamHTTP{
		handler: ctx,
		ctx:     context.Background(),
		method:  method,
	}

	if err = stream.SendHeader(metadata.Pairs("x-pre", "1")); err != nil {
		t.Fatalf("preset SendHeader failed: %v", err)
	}

	if err = mux.invokeResponseStream(stream, &emptypb.Empty{}); err != nil {
		t.Fatalf("invokeResponseStream failed: %v", err)
	}

	body := string(ctx.Response().Body())
	if strings.Count(body, "\n") != 2 {
		t.Fatalf("expected 2 NDJSON lines, got body=%q", body)
	}
	if !strings.Contains(body, "\"msg\":\"a\"") || !strings.Contains(body, "\"msg\":\"b\"") {
		t.Fatalf("stream body missing expected messages: %q", body)
	}
	if got := string(ctx.Response().Header.Peek("x-stream")); got != "header" {
		t.Fatalf("missing streamed header, got=%q", got)
	}
	if got := string(ctx.Response().Header.Peek("x-pre")); got != "1" {
		t.Fatalf("preset header lost, got=%q", got)
	}
	if len(fakeStream.sentReqs) != 1 {
		t.Fatalf("expected 1 request message sent, got=%d", len(fakeStream.sentReqs))
	}
}

func TestInvokeResponseStream_RejectsClientStreamingMode(t *testing.T) {
	mux := NewMux()

	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find input type: %v", err)
	}
	outType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Struct")
	if err != nil {
		t.Fatalf("find output type: %v", err)
	}

	method := &methodWrapper{
		srv:            &serviceWrapper{opts: mux.opts, remoteProxyCli: &fakeClientConn{stream: &fakeClientStream{}}},
		grpcStreamDesc: &grpc.StreamDesc{ServerStreams: true, ClientStreams: true},
		grpcFullMethod: "/test.v1.StreamService/Bidi",
		inputType:      inType,
		outputType:     outType,
	}

	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	stream := &streamHTTP{handler: ctx, ctx: context.Background(), method: method}
	err = mux.invokeResponseStream(stream, &emptypb.Empty{})
	if err == nil {
		t.Fatal("expected error for client-streaming mode, got nil")
	}
	if !strings.Contains(err.Error(), "client-streaming") {
		t.Fatalf("unexpected error: %v", err)
	}
}
