package grpc_entry

import (
	"context"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"net"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xprocess"
	"google.golang.org/grpc"
)

var _ golug_entry.GrpcEntry = (*grpcEntry)(nil)

type grpcEntry struct {
	golug_entry.Entry
	cfg                      Cfg
	server                   *grpc.Server
	handlers                 []interface{}
	opts                     []grpc.ServerOption
	unaryServerInterceptors  []grpc.UnaryServerInterceptor
	streamServerInterceptors []grpc.StreamServerInterceptor
}

func (t *grpcEntry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.Entry.Run().Init())
	xerror.Panic(golug_config.Decode(Name, &t.cfg))

	return nil
}

func (t *grpcEntry) Options() golug_entry.Options { return t.Entry.Run().Options() }

func (t *grpcEntry) Run() golug_entry.RunEntry { return t }

func (t *grpcEntry) UnWrap(fn interface{}) error { return xerror.Wrap(golug_entry.UnWrap(t, fn)) }

func (t *grpcEntry) Register(ss interface{}) {
	if ss == nil {
		xerror.Panic(xerror.New("[ss] should not be nil"))
	}

	t.handlers = append(t.handlers, ss)
}

func (t *grpcEntry) WithUnaryServer(interceptors ...golug_entry.UnaryServerInterceptor) {
	var interceptors1 []grpc.UnaryServerInterceptor
	for i := range interceptors {
		interceptors1 = append(interceptors1, func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			defer xerror.RespErr(&err)
			return handler(interceptors[i](ctx, info), req)
		})
	}

	t.unaryServerInterceptors = append(t.unaryServerInterceptors, interceptors1...)
}

func (t *grpcEntry) WithStreamServer(interceptors ...golug_entry.StreamServerInterceptor) {
	var interceptors1 []grpc.StreamServerInterceptor
	for i := range interceptors {
		interceptors1 = append(interceptors1, func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
			defer xerror.RespErr(&err)

			return handler(srv, &grpc_middleware.WrappedServerStream{
				ServerStream:   ss,
				WrappedContext: interceptors[i](ss, info),
			})
		})
	}

	t.streamServerInterceptors = append(t.streamServerInterceptors, interceptors1...)
}

func (t *grpcEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	// 初始化server
	t.server = grpc.NewServer(
		append(t.opts,
			grpc.ChainUnaryInterceptor(t.unaryServerInterceptors...),
			grpc.ChainStreamInterceptor(t.streamServerInterceptors...))...,
	)

	// 初始化routes
	for i := range t.handlers {
		xerror.Panic(register(t.server, t.handlers[i]))
	}

	cancel := xprocess.Go(func(ctx context.Context) (err error) {
		defer xerror.RespErr(&err)

		ts := xerror.PanicErr(net.Listen("tcp", t.Entry.Run().Options().Addr)).(net.Listener)
		log.Infof("Server [grpc] Listening on %s", ts.Addr().String())
		if err := t.server.Serve(ts); err != nil && err != grpc.ErrServerStopped {
			log.Error(err.Error())
		}
		return nil
	})

	xerror.Panic(dix_run.WithBeforeStop(func(ctx *dix_run.BeforeStopCtx) { xerror.Panic(cancel()) }))

	return nil
}

func (t *grpcEntry) Stop() (err error) {
	defer xerror.RespErr(&err)
	t.server.GracefulStop()
	log.Infof("Server [grpc] Closed OK")
	return nil
}

func newEntry(name string) *grpcEntry {
	ent := &grpcEntry{
		Entry: golug_entry.New(name),
	}
	ent.trace()

	return ent
}

func New(name string) *grpcEntry {
	return newEntry(name)
}
