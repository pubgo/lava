package zrpcbridge

import (
	"context"

	"github.com/pubgo/funk/v2/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/pubgo/lava/v2/pkg/gateway"
	"github.com/pubgo/lava/v2/pkg/zrpc"
)

var _ grpc.ServerStream = (*streamAdapter)(nil)

// streamAdapter adapts zrpc.ServerStream to grpc.ServerStream for gateway dispatch.
type streamAdapter struct {
	zstream    *zrpc.ServerStream
	ctx        context.Context
	fullMethod string

	header  metadata.MD
	trailer metadata.MD
}

func (s *streamAdapter) SetHeader(md metadata.MD) error {
	s.header = metadata.Join(s.header, md)
	return nil
}

func (s *streamAdapter) SendHeader(md metadata.MD) error {
	return s.SetHeader(md)
}

func (s *streamAdapter) SetTrailer(md metadata.MD) {
	s.trailer = metadata.Join(s.trailer, md)
}

func (s *streamAdapter) Context() context.Context {
	return gateway.NewContextWithServerTransportStream(s.ctx, s, s.fullMethod)
}

func (s *streamAdapter) SendMsg(m any) error {
	msg, ok := m.(proto.Message)
	if !ok {
		return errors.New("zrpcbridge: send msg: not a proto.Message")
	}
	return s.zstream.Send(msg)
}

func (s *streamAdapter) RecvMsg(m any) error {
	msg, ok := m.(proto.Message)
	if !ok {
		return errors.New("zrpcbridge: recv msg: not a proto.Message")
	}
	return s.zstream.Recv(msg)
}

// streamUnary captures unary responses from gateway dispatch.
type streamUnary struct {
	ctx        context.Context
	fullMethod string
	response   proto.Message
	header     metadata.MD
	trailer    metadata.MD
}

func (s *streamUnary) SetHeader(md metadata.MD) error {
	s.header = metadata.Join(s.header, md)
	return nil
}

func (s *streamUnary) SendHeader(md metadata.MD) error {
	return s.SetHeader(md)
}

func (s *streamUnary) SetTrailer(md metadata.MD) {
	s.trailer = metadata.Join(s.trailer, md)
}

func (s *streamUnary) Context() context.Context {
	return gateway.NewContextWithServerTransportStream(s.ctx, s, s.fullMethod)
}

func (s *streamUnary) SendMsg(m any) error {
	msg, ok := m.(proto.Message)
	if !ok {
		return errors.New("zrpcbridge: unary send msg: not a proto.Message")
	}
	s.response = msg
	return nil
}

func (s *streamUnary) RecvMsg(any) error {
	return errors.New("zrpcbridge: unary RecvMsg not supported")
}
