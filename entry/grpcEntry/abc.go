package grpcEntry

import (
	"google.golang.org/grpc"

	"github.com/pubgo/lava/entry"

	// grpc log插件加载
	_ "github.com/pubgo/lava/internal/plugins/grpclog"
)

type Entry interface {
	entry.Entry
	Register(handler entry.InitHandler)
	UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor)
	StreamInterceptor(interceptors ...grpc.StreamServerInterceptor)
}
