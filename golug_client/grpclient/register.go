package grpclient

import (
	"sync"

	"github.com/pubgo/xerror/xerror_util"
	"google.golang.org/grpc"
)

var interceptorMap sync.Map

func RegisterUnary(interceptor grpc.UnaryClientInterceptor) {
	interceptorMap.Store(xerror_util.CallerWithFunc(interceptor), interceptor)
}

func RegisterStream(interceptor grpc.StreamClientInterceptor) {
	interceptorMap.Store(xerror_util.CallerWithFunc(interceptor), interceptor)
}
