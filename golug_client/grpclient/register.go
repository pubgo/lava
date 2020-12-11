package grpclient

import (
	"sync"

	"github.com/pubgo/xerror/xerror_util"
	"google.golang.org/grpc"
)

var data sync.Map

func RegisterUnary(interceptor grpc.UnaryClientInterceptor) {
	data.Store(xerror_util.CallerWithFunc(interceptor), interceptor)
}

func RegisterStream(interceptor grpc.StreamClientInterceptor) {
	data.Store(xerror_util.CallerWithFunc(interceptor), interceptor)
}
