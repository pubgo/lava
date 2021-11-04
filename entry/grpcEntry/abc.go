package grpcEntry

import (
	"github.com/pubgo/lava/entry"
	"google.golang.org/grpc"

	// grpc log插件加载
	_ "github.com/pubgo/lava/internal/plugins/grpclog"
)

type Entry interface {
	entry.Entry
	UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor)
	StreamInterceptor(interceptors ...grpc.StreamServerInterceptor)
	Register(handler entry.InitHandler)
}
