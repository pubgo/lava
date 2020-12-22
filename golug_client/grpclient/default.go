package grpclient

import (
	"context"
	"time"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

func GetClient(name string) grpc.ClientConnInterface {
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
	_, ok := connPool.LoadOrStore(name, &grpcPool{})
	if ok {
		xerror.Next().Exit(xerror.Fmt("%s already exists", name))
	}

	cc := createConn(name)
	defer cc.conn.Close()
	return cc.conn
}

func createConn(addr string) *grpcConn {
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
	return &grpcConn{service: addr, conn: cc, updated: time.Now()}
}

func init() {
	xerror.Exit(dix_run.WithBeforeStart(func(ctx *dix_run.BeforeStartCtx) {
		// 服务启动之前, 初始化grpc conn pool
		connPool.Range(func(key, value interface{}) bool {
			name := key.(string)
			pool := value.(*grpcPool)

			// idleNum
			for i := 5; i > 0; i-- {
				cc := createConn(name)
				pool.connList = append(pool.connList, cc)
				pool.connMap.Store(cc, struct{}{})
			}

			return true
		})
	}))
}
