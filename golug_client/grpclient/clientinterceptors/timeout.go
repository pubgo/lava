package clientinterceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func ShrinkDeadline(ctx context.Context, timeout time.Duration) (context.Context, func()) {
	if deadline, ok := ctx.Deadline(); ok {
		leftTime := time.Until(deadline)
		if leftTime < timeout {
			timeout = leftTime
		}
	}

	return context.WithDeadline(ctx, time.Now().Add(timeout))
}

const (
	DefaultUnaryTimeout  = 10 * time.Second
	DefaultStreamTimeout = 10 * time.Second
)

// TimeoutUnary returns a new unary server interceptor for OpenTracing.
func TimeoutUnary(t time.Duration) grpc.UnaryClientInterceptor {
	defaultTimeOut := DefaultUnaryTimeout
	if t <= 0 {
		defaultTimeOut = t
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := ShrinkDeadline(ctx, defaultTimeOut)
		defer cancel()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// TimeoutStream returns a new streaming server interceptor for panic recovery.
func TimeoutStream(t time.Duration) grpc.StreamClientInterceptor {
	defaultTimeOut := DefaultStreamTimeout
	if t <= 0 {
		defaultTimeOut = t
	}

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx, cancel := ShrinkDeadline(ctx, defaultTimeOut)
		defer cancel()
		return streamer(ctx, desc, cc, method, opts...)
	}
}
