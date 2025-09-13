package gateway

import (
	"context"

	"github.com/pubgo/funk/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ grpc.ServerTransportStream = (*serverTransportStream)(nil)

func NewContextWithServerTransportStream(ctx context.Context, s grpc.ServerStream, method string) context.Context {
	assert.If(ctx == nil, "ctx is nil")
	assert.If(s == nil, "server stream is nil")
	assert.If(method == "", "method is nil")
	return grpc.NewContextWithServerTransportStream(ctx, &serverTransportStream{ServerStream: s, method: method})
}

// serverTransportStream wraps grpc.SeverStream to support header/trailers.
type serverTransportStream struct {
	grpc.ServerStream
	method string
}

func (s *serverTransportStream) Method() string { return s.method }
func (s *serverTransportStream) SetHeader(md metadata.MD) error {
	return s.ServerStream.SetHeader(md)
}

func (s *serverTransportStream) SendHeader(md metadata.MD) error {
	return s.ServerStream.SendHeader(md)
}

func (s *serverTransportStream) SetTrailer(md metadata.MD) error {
	s.ServerStream.SetTrailer(md)
	return nil
}
