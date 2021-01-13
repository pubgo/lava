package client

import (
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc"
)

// middleware for grpc unary calls
var defaultUnaryInterceptor = []grpc.UnaryClientInterceptor{grpc_opentracing.UnaryClientInterceptor()}

// middleware for grpc stream calls
var defaultStreamInterceptor = []grpc.StreamClientInterceptor{grpc_opentracing.StreamClientInterceptor()}

func getDialOption() []grpc.DialOption {
	dialOpts := append(defaultDialOpts, grpc.WithChainUnaryInterceptor(defaultUnaryInterceptor...),
		grpc.WithChainStreamInterceptor(defaultStreamInterceptor...),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(DefaultMaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(DefaultMaxSendMsgSize)))
	return dialOpts
}

func AppendUnaryInterceptor(unaryInterceptor ...grpc.UnaryClientInterceptor) {
	defaultUnaryInterceptor = append(defaultUnaryInterceptor, unaryInterceptor...)
}

func AppendStreamInterceptor(streamInterceptor ...grpc.StreamClientInterceptor) {
	defaultStreamInterceptor = append(defaultStreamInterceptor, streamInterceptor...)
}

func WithStreamInterceptors(streamInterceptor ...grpc.StreamClientInterceptor) grpc.DialOption {
	AppendStreamInterceptor(streamInterceptor...)
	return grpc.EmptyDialOption{}
}

func WithUnaryInterceptors(unaryInterceptor ...grpc.UnaryClientInterceptor) grpc.DialOption {
	AppendUnaryInterceptor(unaryInterceptor...)
	return grpc.EmptyDialOption{}
}
