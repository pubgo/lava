package gateway

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/emptypb"
)

type proxyFakeClientStream struct {
	header metadata.MD
	msgs   []proto.Message
	idx    int
}

func (p *proxyFakeClientStream) Header() (metadata.MD, error) { return p.header, nil }
func (p *proxyFakeClientStream) Trailer() metadata.MD         { return nil }
func (p *proxyFakeClientStream) CloseSend() error             { return nil }
func (p *proxyFakeClientStream) Context() context.Context     { return context.Background() }
func (p *proxyFakeClientStream) SendMsg(any) error            { return nil }
func (p *proxyFakeClientStream) RecvMsg(m any) error {
	if p.idx >= len(p.msgs) {
		return io.EOF
	}
	pm, ok := m.(proto.Message)
	if !ok {
		return io.ErrUnexpectedEOF
	}
	b, err := proto.Marshal(p.msgs[p.idx])
	if err != nil {
		return err
	}
	p.idx++
	return proto.Unmarshal(b, pm)
}

type proxyFakeServerStream struct {
	sentMessages int
	sentHeader   metadata.MD
}

func (p *proxyFakeServerStream) SetHeader(metadata.MD) error { return nil }
func (p *proxyFakeServerStream) SendHeader(md metadata.MD) error {
	if p.sentHeader != nil {
		return errors.New("headers already sent")
	}
	p.sentHeader = md
	return nil
}
func (p *proxyFakeServerStream) SetTrailer(metadata.MD)   {}
func (p *proxyFakeServerStream) Context() context.Context { return context.Background() }
func (p *proxyFakeServerStream) SendMsg(any) error        { p.sentMessages++; return nil }
func (p *proxyFakeServerStream) RecvMsg(any) error        { return io.EOF }

func TestPumpBackendToFrontend_PropagatesHeader(t *testing.T) {
	outType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find output type: %v", err)
	}

	src := &proxyFakeClientStream{
		header: metadata.Pairs("x-test", "1"),
		msgs:   []proto.Message{&emptypb.Empty{}},
	}
	dst := &proxyFakeServerStream{}

	errCh := pumpBackendToFrontend(outType, src, dst, true)
	if got := <-errCh; got != io.EOF {
		t.Fatalf("expected io.EOF, got %v", got)
	}
	if dst.sentMessages != 1 {
		t.Fatalf("expected 1 forwarded message, got %d", dst.sentMessages)
	}
	if got := dst.sentHeader.Get("x-test"); len(got) != 1 || got[0] != "1" {
		t.Fatalf("missing propagated header, got=%v", dst.sentHeader)
	}
}

func TestPumpBackendToFrontend_IgnoresDuplicateHeaderError(t *testing.T) {
	outType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find output type: %v", err)
	}

	src := &proxyFakeClientStream{
		header: metadata.Pairs("x-test", "1"),
		msgs:   []proto.Message{&emptypb.Empty{}},
	}
	dst := &proxyFakeServerStream{sentHeader: metadata.Pairs("already", "1")}

	errCh := pumpBackendToFrontend(outType, src, dst, true)
	if got := <-errCh; got != io.EOF {
		t.Fatalf("expected io.EOF, got %v", got)
	}
	if dst.sentMessages != 1 {
		t.Fatalf("expected 1 forwarded message, got %d", dst.sentMessages)
	}
}

func TestIsDuplicateHeaderError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "headers already sent", err: errors.New("headers already sent"), want: true},
		{name: "SendHeader called multiple times", err: errors.New("SendHeader called multiple times"), want: true},
		{name: "other", err: errors.New("connection reset by peer"), want: false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDuplicateHeaderError(tt.err); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDispatchBidi_CloseSendErrorStillSetsTrailer(t *testing.T) {
	// When the client half finishes and CloseSend fails, the dispatcher must still
	// attach backend trailers so the frontend sees grpc-status / custom trailers.
	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find input type: %v", err)
	}
	outType := inType

	closeErr := errors.New("close-send failed")
	backendStream := &bidiCloseSendClientStream{
		trailer:  metadata.Pairs("grpc-status", "0", "x-trail", "ok"),
		closeErr: closeErr,
		closed:   make(chan struct{}),
	}
	frontend := &bidiCloseSendServerStream{}
	backend := &bidiCloseSendBackend{stream: backendStream}

	op := &Operation{
		FullMethod: "/test.v1.Echo/Bidi",
		InputType:  inType,
		OutputType: outType,
		StreamDesc: &grpc.StreamDesc{ClientStreams: true, ServerStreams: true},
	}

	d := &Dispatcher{}
	err = d.dispatchBidi(context.Background(), backend, frontend, op, dispatchConfig{})
	if !errors.Is(err, closeErr) {
		t.Fatalf("want CloseSend error wrapped, got %v", err)
	}
	if got := frontend.trailer.Get("x-trail"); len(got) != 1 || got[0] != "ok" {
		t.Fatalf("trailer not set after CloseSend failure, got=%v", frontend.trailer)
	}
	if got := frontend.trailer.Get("grpc-status"); len(got) != 1 || got[0] != "0" {
		t.Fatalf("grpc-status trailer missing, got=%v", frontend.trailer)
	}
}

type bidiCloseSendBackend struct {
	stream grpc.ClientStream
}

func (b *bidiCloseSendBackend) Invoke(context.Context, string, any, any, ...grpc.CallOption) error {
	return errors.New("unexpected Invoke")
}
func (b *bidiCloseSendBackend) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return b.stream, nil
}

type bidiCloseSendClientStream struct {
	trailer  metadata.MD
	closeErr error
	closed   chan struct{}
	once     sync.Once
}

func (s *bidiCloseSendClientStream) Header() (metadata.MD, error) { return nil, nil }
func (s *bidiCloseSendClientStream) Trailer() metadata.MD         { return s.trailer }
func (s *bidiCloseSendClientStream) CloseSend() error {
	s.once.Do(func() { close(s.closed) })
	return s.closeErr
}
func (s *bidiCloseSendClientStream) Context() context.Context { return context.Background() }
func (s *bidiCloseSendClientStream) SendMsg(any) error        { return nil }
func (s *bidiCloseSendClientStream) RecvMsg(any) error {
	// Block until CloseSend so the s2c path runs first (otherwise c2s EOF can win
	// the select and skip CloseSend entirely).
	<-s.closed
	return io.EOF
}

type bidiCloseSendServerStream struct {
	trailer metadata.MD
}

func (s *bidiCloseSendServerStream) SetHeader(metadata.MD) error  { return nil }
func (s *bidiCloseSendServerStream) SendHeader(metadata.MD) error { return nil }
func (s *bidiCloseSendServerStream) SetTrailer(md metadata.MD)    { s.trailer = md }
func (s *bidiCloseSendServerStream) Context() context.Context     { return context.Background() }
func (s *bidiCloseSendServerStream) SendMsg(any) error            { return nil }
func (s *bidiCloseSendServerStream) RecvMsg(any) error            { return io.EOF }

func TestUseBackendUnaryInterceptor_CoversProxy(t *testing.T) {
	mux := NewMux()
	var hits atomic.Int32
	mux.UseBackendUnaryInterceptor(func(ctx context.Context, method string, req, reply any, invoker func(context.Context, string, any, any, ...grpc.CallOption) error, opts ...grpc.CallOption) error {
		hits.Add(1)
		return invoker(ctx, method, req, reply, opts...)
	})

	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find type: %v", err)
	}
	method := &methodWrapper{
		srv: &serviceWrapper{
			opts:           mux.opts,
			remoteProxyCli: &fakeClientConn{},
		},
		grpcFullMethod: "/test.v1.Echo/Ping",
		inputType:      inType,
		outputType:     inType,
	}
	mux.opts.handlers[method.grpcFullMethod] = method

	if err = mux.Invoke(context.Background(), method.grpcFullMethod, &emptypb.Empty{}, &emptypb.Empty{}); err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("backend interceptor hits=%d want 1", hits.Load())
	}
}
