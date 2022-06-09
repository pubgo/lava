package grpc_builder

import (
	"github.com/pubgo/lava/core/runmode"
	"time"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Builder struct {
	srv                *grpc.Server
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
}

func (t *Builder) UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) {
	t.unaryInterceptors = append(t.unaryInterceptors, interceptors...)
}

func (t *Builder) StreamInterceptor(interceptors ...grpc.StreamServerInterceptor) {
	t.streamInterceptors = append(t.streamInterceptors, interceptors...)
}

func (t *Builder) Get() *grpc.Server {
	if t.srv == nil {
		panic("srv is nil, please init grpc server")
	}

	return t.srv
}

func (t *Builder) BuildOpts(cfg *Cfg) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.MaxSendMsgSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
			PermitWithoutStream: true,            // Allow pings even when there are no active streams
		}),
	}
}

func (t *Builder) Build(cfg *Cfg) (err error) {
	defer xerror.RecoverErr(&err)

	opts := t.BuildOpts(cfg)
	opts = append(opts, grpc.ChainUnaryInterceptor(t.unaryInterceptors...))
	opts = append(opts, grpc.ChainStreamInterceptor(t.streamInterceptors...))
	t.srv = grpc.NewServer(opts...)

	EnableReflection(t.srv)
	EnableHealth("", t.srv)
	if runmode.IsDev() || runmode.IsTest() {
		EnableDebug(t.srv)
	}

	return nil
}

func New() Builder { return Builder{} }
