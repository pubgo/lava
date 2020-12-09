package grpc_client

import (
	"context"

	"github.com/pubgo/golug/golug_consts"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

func GetCfg() Cfg {
	return cfg
}

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

}

func GetClient(names ...string) (grpc.ClientConnInterface, error) {
	var name = golug_consts.Default
	if len(names) > 0 {
		name = names[0]
	}

	val, ok := clientM.Load(name)
	if !ok {
		return nil, xerror.Fmt("%s not found", name)
	}

	return val.(grpc.ClientConnInterface), nil
}

func initClient(name string, cfg ClientCfg) {

	clientM.Store(name, nil)
}
