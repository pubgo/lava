package grpclient

import (
	"context"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/pubgo/xerror"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
)

var clientM sync.Map
var connPool sync.Map

// DefaultMaxRecvMsgSize maximum message that client can receive
// (4 MB).
var DefaultMaxRecvMsgSize = 1024 * 1024 * 4

var maxConnRef = uint32(50)

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

func unaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// select grpc conn from grpc client pool
	service := cc.Target()
	cc1 := selectConn(service)
	defer releaseConn(cc1)
	return invoker(ctx, method, req, reply, cc1.conn, opts...)
}

func streamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	// select grpc conn from grpc client pool
	service := cc.Target()
	cc1 := selectConn(service)
	defer releaseConn(cc1)
	return streamer(ctx, desc, cc1.conn, method, opts...)
}

func isConnInvalid(conn *grpcConn) bool {
	return conn.closed ||
		conn.conn.GetState() == connectivity.Shutdown ||
		conn.conn.GetState() == connectivity.TransientFailure
}

// selectConn get new conn from clientPools by conn.Target()
// conn.Target() is equal to serviceName
func selectConn(service string) *grpcConn {
	val, ok := connPool.Load(service)
	if !ok {
		xerror.Next().Panic(xerror.Fmt("%s not found", service))
	}

	var conn *grpcConn
	defer func() { conn.connRef.Inc() }()

	var isValid bool
	pool := val.(*grpcPool)
	pool.connMap.Range(func(key, value interface{}) bool {
		conn = key.(*grpcConn)
		if isConnInvalid(conn) {
			return true
		}

		isValid = true
		return false
	})

	isOkRef := conn.connRef.Load() <= maxConnRef
	if isValid && isOkRef {
		return conn
	}

	// 创建新的grpc conn
	cc := createConn(service)

	if !isValid {
		conn.closed = false
		conn.conn = cc
	}

	if !isOkRef {
		conn = &grpcConn{conn: cc}
		pool.connList = append(pool.connList, conn)
		pool.connMap.Store(conn, struct{}{})
	}

	return conn
}

// when grpc call is finished, release the grpc Conn
func releaseConn(conn *grpcConn) {
	conn.connRef.Dec()
}

type grpcPool struct {
	// 最大连接引用数
	maxConnRef int32

	// virtual grpc conn
	// vConn  *grpc.ClientConn
	closed bool

	//count(streams) this clientPool is using now.
	//it should < size*maxConcurrentStreams forever
	poolRef *atomic.Uint32

	// 管理当前可选的连接
	connList []*grpcConn
	connMap  sync.Map
}

type grpcConn struct {
	connRef atomic.Uint32
	conn    *grpc.ClientConn
	closed  bool
}
