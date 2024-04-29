package gateway

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ grpc.ServerTransportStream = (*serverTransportStream)(nil)

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
