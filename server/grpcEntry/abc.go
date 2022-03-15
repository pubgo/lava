package grpcEntry

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/server"

	// grpc log插件加载
	_ "github.com/pubgo/lava/internal/plugins/grpclog"
)

type Entry interface {
	server.Entry
	grpc.ServiceRegistrar
	Mux() *runtime.ServeMux
	Conn() grpc.ClientConnInterface
	UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor)
	StreamInterceptor(interceptors ...grpc.StreamServerInterceptor)
}
