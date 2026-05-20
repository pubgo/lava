package gateway

import (
	"context"
	"errors"
	"io"
	"testing"

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
}

func (p *proxyFakeServerStream) SetHeader(metadata.MD) error { return nil }
func (p *proxyFakeServerStream) SendHeader(metadata.MD) error {
	return errors.New("headers already sent")
}
func (p *proxyFakeServerStream) SetTrailer(metadata.MD)   {}
func (p *proxyFakeServerStream) Context() context.Context { return context.Background() }
func (p *proxyFakeServerStream) SendMsg(any) error        { p.sentMessages++; return nil }
func (p *proxyFakeServerStream) RecvMsg(any) error        { return io.EOF }

func TestForwardClientToServer_IgnoresDuplicateHeaderError(t *testing.T) {
	outType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find output type: %v", err)
	}

	src := &proxyFakeClientStream{
		header: metadata.Pairs("x-test", "1"),
		msgs:   []proto.Message{&emptypb.Empty{}},
	}
	dst := &proxyFakeServerStream{}

	errCh := forwardClientToServer(outType, src, dst)
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
