package golug_middleware

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

func GetDefaultServerOpts() []grpc.ServerOption {

	_ = []grpc.ServerOption{
		grpc.MaxRecvMsgSize(1),
		grpc.MaxSendMsgSize(1),
		grpc.UnknownServiceHandler(nil),
		grpc.Creds(credentials.NewTLS(nil)),
	}

	return []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{}),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(nil)),
			grpc_opentracing.UnaryServerInterceptor(),
			ratelimit.UnaryServerInterceptor(nil),
			grpc_auth.UnaryServerInterceptor(nil),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(nil)),
			grpc_opentracing.StreamServerInterceptor(),
			ratelimit.StreamServerInterceptor(nil),
			grpc_auth.StreamServerInterceptor(nil),
		)),
	}
}
