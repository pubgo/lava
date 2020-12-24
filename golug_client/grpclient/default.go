package grpclient

import (
	"context"
	"sync"
	"time"

	"github.com/pubgo/golug/pkg/golug_utils"
	"github.com/pubgo/xerror"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
)

func GetClient(name string) grpc.ClientConnInterface {
	return nil
}

type Client interface {
	Name() string
}

type baseClient struct {
	name string
}

func (t baseClient) Name() string { return t.name }

func Init(name string) Client {
	_, ok := connPool.LoadOrStore(name, &grpcPool{})
	if ok {
		xerror.Next().Exit(xerror.Fmt("%s already exists", name))
	}
	return baseClient{name: name}
}

func unaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// select grpc conn from grpc client pool
	cc1 := selectConn(cc.Target())
	defer releaseConn(cc1)
	return invoker(ctx, method, req, reply, cc1.conn, opts...)
}

func streamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	// select grpc conn from grpc client pool
	cc1 := selectConn(cc.Target())
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
	conn = pool.createConn()

	if !isValid {
		conn.closed = false
	}

	if !isOkRef {
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
	cfg ClientCfg
	// 最大连接引用数
	maxConnRef int32

	addr string

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

func (t *grpcPool) handleOpts() []grpc.DialOption {
	var cfg = t.cfg

	var opts []grpc.DialOption
	if cfg.Insecure {
		opts = append(opts, grpc.WithInsecure())
	}
	if cfg.Block {
		opts = append(opts, grpc.WithBlock())
	}

	var s keepalive.ClientParameters
	golug_utils.Mergo(&s, cfg.ClientParameters)
	opts = append(opts, grpc.WithKeepaliveParams(s))

	var cos []grpc.CallOption
	cos = append(cos, grpc.MaxCallSendMsgSize(cfg.Call.MaxCallSendMsgSize))
	cos = append(cos, grpc.MaxCallRecvMsgSize(cfg.Call.MaxCallRecvMsgSize))
	opts = append(opts, grpc.WithDefaultCallOptions(cos...))

	return opts
}

func (t *grpcPool) createConn() *grpcConn {
	var addr = t.addr
	var defaultUnaryInterceptor []grpc.UnaryClientInterceptor
	var defaultStreamInterceptor []grpc.StreamClientInterceptor

	interceptorMap.Range(func(key, value interface{}) bool {
		switch value := value.(type) {
		case grpc.UnaryClientInterceptor:
			defaultUnaryInterceptor = append(defaultUnaryInterceptor, value)
		case grpc.StreamClientInterceptor:
			defaultStreamInterceptor = append(defaultStreamInterceptor, value)
		}
		return true
	})

	var opts = []grpc.DialOption{grpc.WithUnaryInterceptor(unaryInterceptor), grpc.WithStreamInterceptor(streamInterceptor)}
	opts = append(opts, grpc.WithChainUnaryInterceptor(defaultUnaryInterceptor...))
	opts = append(opts, grpc.WithChainStreamInterceptor(defaultStreamInterceptor...))
	opts = append(opts, t.handleOpts()...)

	ctx, cancel := context.WithTimeout(context.Background(), t.cfg.DialTimeout)
	defer cancel()
	cc, err := grpc.DialContext(ctx, addr, opts...)
	xerror.Next().Panic(err)
	return &grpcConn{service: addr, conn: cc, updated: time.Now()}
}

type grpcConn struct {
	updated time.Time
	service string
	connRef atomic.Uint32
	conn    *grpc.ClientConn
	closed  bool
}
