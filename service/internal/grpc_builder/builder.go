package grpc_builder

import (
	"time"

	"github.com/fullstorydev/grpchan"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/pubgo/lava/runtime"
)

type Builder struct {
	grpchan.HandlerMap
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

func (t *Builder) BuildOpts(cfg Cfg) []grpc.ServerOption {
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

func (t *Builder) Build(cfg Cfg) (err error) {
	defer xerror.RespErr(&err)

	opts := t.BuildOpts(cfg)
	opts = append(opts, grpc.ChainUnaryInterceptor(t.unaryInterceptors...))
	opts = append(opts, grpc.ChainStreamInterceptor(t.streamInterceptors...))
	t.srv = grpc.NewServer(opts...)

	t.HandlerMap.ForEach(func(desc *grpc.ServiceDesc, svr interface{}) { t.srv.RegisterService(desc, svr) })

	EnableReflection(t.srv)
	EnableHealth("", t.srv)
	if runtime.IsDev() || runtime.IsTest() {
		EnableDebug(t.srv)
	}

	return nil
}

func New() Builder { return Builder{HandlerMap: grpchan.HandlerMap{}} }
