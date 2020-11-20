package golug_middleware

import (
	"crypto/tls"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func GetDefaultServerOpts() []grpc.ServerOption {
	gopts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
		grpc.UnknownServiceHandler(g.handler),
		grpc.Creds(credentials.NewTLS(v.(*tls.Config))),
	}

	return []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(defaultRecoveryHandler)),
			TimeoutUnaryServerInterceptor(defaultUnaryTimeout),
			grpc_opentracing.UnaryServerInterceptor(),
			ratelimit.UnaryServerInterceptor(defaultRateLimiter),
			grpc_auth.UnaryServerInterceptor(defaultAuthFunc),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(defaultRecoveryHandler)),
			TimeoutStreamServerInterceptor(defaultStreamTimeout),
			grpc_opentracing.StreamServerInterceptor(),
			ratelimit.StreamServerInterceptor(defaultRateLimiter),
			grpc_auth.StreamServerInterceptor(defaultAuthFunc),
		)),
	}
}
