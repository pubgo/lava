package gateway

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ grpc.ServerStream = (*serverProxyStream)(nil)

type serverProxyStream struct {
	cli grpc.ClientStream
}

func (s serverProxyStream) SetHeader(md metadata.MD) error {
	//TODO implement me
	panic("implement me")
}

func (s serverProxyStream) SendHeader(md metadata.MD) error {
	//TODO implement me
	panic("implement me")
}

func (s serverProxyStream) SetTrailer(md metadata.MD) {
	//TODO implement me
	panic("implement me")
}

func (s serverProxyStream) Context() context.Context {
	//TODO implement me
	panic("implement me")
}

func (s serverProxyStream) SendMsg(m any) error {
	//TODO implement me
	panic("implement me")
}

func (s serverProxyStream) RecvMsg(m any) error {
	//TODO implement me
	panic("implement me")
}
