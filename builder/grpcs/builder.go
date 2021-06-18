package grpcs

import (
	grpcMid "github.com/grpc-ecosystem/go-grpc-middleware"
	opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"sync"
	"time"
)

type Builder struct {
	name                     string
	mu                       sync.Mutex
	srv                      *grpc.Server
	opts                     []grpc.ServerOption
	unaryServerInterceptors  []grpc.UnaryServerInterceptor
	streamServerInterceptors []grpc.StreamServerInterceptor
}

func (t *Builder) UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.unaryServerInterceptors = append(t.unaryServerInterceptors, interceptors...)
}

func (t *Builder) StreamInterceptor(interceptors ...grpc.StreamServerInterceptor) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.streamServerInterceptors = append(t.streamServerInterceptors, interceptors...)
}

func (t *Builder) Get() *grpc.Server {
	if t.srv == nil {
		panic("srv is nil, please init grpc server")
	}

	return t.srv
}

func (t *Builder) Init(opts ...grpc.ServerOption) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.opts = append(t.opts, opts...)
}

func (t *Builder) BuildOpts(cfg *Cfg) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.MaxSendMsgSize),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
			PermitWithoutStream: true,            // Allow pings even when there are no active streams
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{}),
	}
}

func (t *Builder) Build(cfg *Cfg, cb ...func()) (err error) {
	defer xerror.RespErr(&err)

	opts := t.BuildOpts(cfg)

	unaryInterceptorList := append([]grpc.UnaryServerInterceptor{
		opentracing.UnaryServerInterceptor(),
	}, t.unaryServerInterceptors...)

	streamInterceptorList := append([]grpc.StreamServerInterceptor{
		opentracing.StreamServerInterceptor(),
	}, t.streamServerInterceptors...)

	opts = append(opts, grpc.UnaryInterceptor(grpcMid.ChainUnaryServer(unaryInterceptorList...)))
	opts = append(opts, grpc.StreamInterceptor(grpcMid.ChainStreamServer(streamInterceptorList...)))
	opts = append(opts, t.opts...)
	srv := grpc.NewServer(opts...)

	EnableReflection(srv)
	EnableHealth(t.name, srv)
	if runenv.IsDev() || runenv.IsTest() {
		EnableDebug(srv)
	}

	t.srv = srv

	if len(cb) > 0 {
		cb[0]()
	}

	return nil
}

func New(name string) Builder { return Builder{name: name} }
