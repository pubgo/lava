package grpclient

import (
	"context"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

func GetClient1(name string) grpc.ClientConnInterface {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultClientDialTimeout)
	defer cancel()
	_, _ = grpc.DialContext(ctx, name,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(ka),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(DefaultMaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(DefaultMaxSendMsgSize)),
		grpc.WithChainUnaryInterceptor(defaultUnaryInterceptor),
		grpc.WithChainStreamInterceptor(defaultStreamInterceptor))
	return nil
}

func GetClient(name string) grpc.ClientConnInterface {
	val, ok := clientM.Load(name)
	if !ok {
		xerror.Next().Panic(xerror.Fmt("%s not found", name))
	}

	return val.(grpc.ClientConnInterface)
}

func initClient(name string, cfg ClientCfg) {
	clientM.Store(name, nil)
}
