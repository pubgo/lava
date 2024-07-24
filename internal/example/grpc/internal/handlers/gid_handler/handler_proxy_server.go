package gid_handler

import (
	"context"
	"fmt"

	"github.com/pubgo/lava/internal/example/grpc/pkg/proto/gidpb"
	"github.com/pubgo/lava/lava"
	"google.golang.org/grpc"
)

var (
	_ lava.GrpcRouter     = (*IdProxyServer)(nil)
	_ gidpb.IdProxyServer = (*IdProxyServer)(nil)
)

func NewIdProxyServer() lava.GrpcRouter {
	return &IdProxyServer{}
}

type IdProxyServer struct {
}

func (i IdProxyServer) Echo(ctx context.Context, req *gidpb.EchoReq) (*gidpb.EchoRsp, error) {
	fmt.Println("get echo request", req.String())
	return &gidpb.EchoRsp{Hello: req.Hello}, nil
}

func (i IdProxyServer) Middlewares() []lava.Middleware {
	//TODO implement me
	return nil
}

func (i IdProxyServer) ServiceDesc() *grpc.ServiceDesc { return &gidpb.IdProxy_ServiceDesc }
