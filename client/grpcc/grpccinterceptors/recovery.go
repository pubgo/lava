package grpccinterceptors

import (
	"context"

	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RecoveryUnary(f grpcRecovery.RecoveryHandlerFuncContext) grpc.UnaryClientInterceptor {
	if f == nil {
		panic("[f] should not be nil")
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = f(ctx, r)
			}
		}()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func RecoveryStream(f grpcRecovery.RecoveryHandlerFuncContext) grpc.StreamClientInterceptor {
	if f == nil {
		panic("[f] should not be nil")
	}

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (_ grpc.ClientStream, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = f(ctx, r)
			}
		}()

		return streamer(ctx, desc, cc, method, opts...)
	}
}

func DefaultRecovery() grpcRecovery.RecoveryHandlerFuncContext {
	return func(ctx context.Context, p interface{}) (err error) {
		return status.Errorf(codes.Internal, "[grpc] client recovery error, err: %v", p)
	}
}
