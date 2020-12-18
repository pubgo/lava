package grpclient

import (
	"context"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

func GetClient1(name string) grpc.ClientConnInterface {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultClientDialTimeout)
	defer cancel()
	cc, err := grpc.DialContext(ctx, name,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(ka),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(DefaultMaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(DefaultMaxSendMsgSize)),
		grpc.WithChainUnaryInterceptor(unaryInterceptor, defaultUnaryInterceptor),
		grpc.WithChainStreamInterceptor(streamInterceptor, defaultStreamInterceptor))
	xerror.Panic(err)
	return cc
}

func Init(name string) grpc.ClientConnInterface {
	_, ok := clientM.LoadOrStore(name, &grpcPool{})
	if ok {
		xerror.Next().Exit(xerror.Fmt("%s already exists", name))
	}

	cc := createConn(name)
	defer cc.Close()
	return cc
}

func createConn(addr string) *grpc.ClientConn {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultClientDialTimeout)
	defer cancel()
	cc, err := grpc.DialContext(ctx, addr,
		grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithUnaryInterceptor(unaryInterceptor),
		grpc.WithStreamInterceptor(streamInterceptor),
		grpc.WithChainUnaryInterceptor(defaultUnaryInterceptor),
		grpc.WithChainStreamInterceptor(defaultStreamInterceptor),
	)
	xerror.Next().Panic(err)
	return cc
}

func init() {
	xerror.Exit(dix_run.WithBeforeStart(func(ctx *dix_run.BeforeStartCtx) {
		// 服务启动之前, 初始化grpc conn pool
		connPool.Range(func(key, value interface{}) bool {
			name := key.(string)
			pool := value.(*grpcPool)

			// idleNum
			for i := 5; i > 0; i-- {
				cc := &grpcConn{conn: createConn(name)}
				pool.connList = append(pool.connList, cc)
				pool.connMap.Store(cc, struct{}{})
			}

			return true
		})
	}))
}
