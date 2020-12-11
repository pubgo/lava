package grpclient

import (
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc/keepalive"
)

var Name = "grpc_client"
var cfg Cfg
var clientM sync.Map

// DefaultMaxRecvMsgSize maximum message that client can receive
// (4 MB).
var DefaultMaxRecvMsgSize = 1024 * 1024 * 4

// DefaultMaxSendMsgSize maximum message that client can send
// (4 MB).
var DefaultMaxSendMsgSize = 1024 * 1024 * 4

var DefaultClientDialTimeout = 2 * time.Second
var ka = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             2 * time.Second,  // wait 2 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

// middleware for grpc unary calls
var defaultUnaryInterceptor = grpc_middleware.ChainUnaryClient(grpc_opentracing.UnaryClientInterceptor())

// middleware for grpc stream calls
var defaultStreamInterceptor = grpc_middleware.ChainStreamClient(grpc_opentracing.StreamClientInterceptor())
