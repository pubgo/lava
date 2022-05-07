package main

import (
	"context"
	"github.com/trusch/grpc-proxy/proxy"
	"github.com/trusch/grpc-proxy/proxy/codec"

	"google.golang.org/grpc/metadata"
	"net"

	_ "github.com/trusch/grpc-proxy/proxy"
	_ "github.com/trusch/grpc-proxy/proxy/codec"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

func main() {
	var conn, err = grpc.Dial("localhost:8080", grpc.WithBlock(), grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.CallContentSubtype((&codec.Proxy{}).Name())))
	xerror.Exit(err)
	defer conn.Close()

	server := grpc.NewServer(grpc.UnknownServiceHandler(proxy.TransparentHandler(func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		outCtx := metadata.NewOutgoingContext(ctx, md.Copy())
		return outCtx, conn, nil
	})))
	lis, err := net.Listen("tcp", ":8081")
	xerror.Exit(err)
	xerror.Exit(server.Serve(lis))
}
