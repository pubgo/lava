package gateway

import (
	"context"

	"github.com/pubgo/funk/v2/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/pubgo/lava/v2/pkg/zrpc"
)

var _ grpc.ServerStream = (*streamZrpc)(nil)

// streamZrpc adapts a zrpc.ServerStream to grpc.ServerStream for the Dispatcher.
type streamZrpc struct {
	zstream *zrpc.ServerStream
	ctx     context.Context
	method  *methodWrapper

	header  metadata.MD
	trailer metadata.MD
}

func (s *streamZrpc) SetHeader(md metadata.MD) error {
	s.header = metadata.Join(s.header, md)
	return nil
}

func (s *streamZrpc) SendHeader(md metadata.MD) error {
	return s.SetHeader(md)
}

func (s *streamZrpc) SetTrailer(md metadata.MD) {
	s.trailer = metadata.Join(s.trailer, md)
}

func (s *streamZrpc) Context() context.Context {
	if s.method == nil {
		return s.ctx
	}
	return NewContextWithServerTransportStream(s.ctx, s, s.method.grpcFullMethod)
}

func (s *streamZrpc) SendMsg(m any) error {
	msg, ok := m.(proto.Message)
	if !ok {
		return errors.New("stream zrpc send msg: not a proto.Message")
	}
	return s.zstream.Send(msg)
}

func (s *streamZrpc) RecvMsg(m any) error {
	msg, ok := m.(proto.Message)
	if !ok {
		return errors.New("stream zrpc recv msg: not a proto.Message")
	}
	return s.zstream.Recv(msg)
}

// streamZrpcUnary captures unary responses from the Dispatcher.
type streamZrpcUnary struct {
	ctx      context.Context
	method   *methodWrapper
	response proto.Message
	header   metadata.MD
	trailer  metadata.MD
}

func (s *streamZrpcUnary) SetHeader(md metadata.MD) error {
	s.header = metadata.Join(s.header, md)
	return nil
}

func (s *streamZrpcUnary) SendHeader(md metadata.MD) error {
	return s.SetHeader(md)
}

func (s *streamZrpcUnary) SetTrailer(md metadata.MD) {
	s.trailer = metadata.Join(s.trailer, md)
}

func (s *streamZrpcUnary) Context() context.Context {
	if s.method == nil {
		return s.ctx
	}
	return NewContextWithServerTransportStream(s.ctx, s, s.method.grpcFullMethod)
}

func (s *streamZrpcUnary) SendMsg(m any) error {
	msg, ok := m.(proto.Message)
	if !ok {
		return errors.New("stream zrpc unary send msg: not a proto.Message")
	}
	s.response = msg
	return nil
}

func (s *streamZrpcUnary) RecvMsg(any) error {
	return errors.New("stream zrpc unary: RecvMsg not supported")
}
