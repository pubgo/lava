package golug_grpc

import (
	"context"
	"github.com/pubgo/golug/internal/golug_util"
	"net"
	"time"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/golug_base"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var _ Entry = (*grpcEntry)(nil)

type grpcEntry struct {
	golug_entry.Entry
	cfg                      Cfg
	server                   *grpc.Server
	handlers                 []interface{}
	opts                     []grpc.ServerOption
	unaryServerInterceptors  []grpc.UnaryServerInterceptor
	streamServerInterceptors []grpc.StreamServerInterceptor
}

func (t *grpcEntry) UnaryServer(interceptors ...grpc.UnaryServerInterceptor) {
	t.unaryServerInterceptors = append(t.unaryServerInterceptors, interceptors...)
}

func (t *grpcEntry) StreamServer(interceptors ...grpc.StreamServerInterceptor) {
	t.streamServerInterceptors = append(t.streamServerInterceptors, interceptors...)
}

func (t *grpcEntry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.Entry.Run().Init())
	golug_config.Decode(Name, &t.cfg)
	return nil
}

func (t *grpcEntry) Options() golug_entry.Options { return t.Entry.Run().Options() }

func (t *grpcEntry) Run() golug_entry.RunEntry { return t }

func (t *grpcEntry) UnWrap(fn interface{}) { xerror.Next().Panic(golug_util.UnWrap(t, fn)) }

func (t *grpcEntry) Register(ss interface{}, opts ...Option) {
	if ss == nil {
		xerror.Panic(xerror.New("[ss] should not be nil"))
	}

	t.handlers = append(t.handlers, ss)
}

func (t *grpcEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	// 初始化server
	t.server = grpc.NewServer(append(
		t.opts,
		grpc.ChainUnaryInterceptor(t.unaryServerInterceptors...),
		grpc.ChainStreamInterceptor(t.streamServerInterceptors...))...)

	// 初始化routes
	for i := range t.handlers {
		xerror.Panic(register(t.server, t.handlers[i]))
	}

	// 方便grpcurl调用和调试
	reflection.Register(t.server)

	cancel := xprocess.GoDelay(time.Second, func(ctx context.Context) {
		defer xerror.Resp(func(err xerror.XErr) {
			xlog.Error("grpcEntry.Start handle error", xlog.Any("err", err))
		})

		ts := xerror.PanicErr(net.Listen("tcp", t.Options().Addr)).(net.Listener)
		xlog.Infof("Server [grpc] Listening on %s", ts.Addr().String())
		if err := t.server.Serve(ts); err != nil && err != grpc.ErrServerStopped {
			xlog.Error(err.Error())
		}
		return
	})

	xerror.Panic(dix_run.WithBeforeStop(func(ctx *dix_run.BeforeStopCtx) { cancel() }))

	return nil
}

func (t *grpcEntry) Stop() (err error) {
	defer xerror.RespErr(&err)
	t.server.GracefulStop()
	xlog.Infof("Server [grpc] Closed OK")
	return nil
}

func newEntry(name string, cfg interface{}) *grpcEntry {
	ent := &grpcEntry{
		Entry: golug_base.New(name, cfg),
	}
	ent.trace()

	return ent
}

func New(name string, cfg interface{}) *grpcEntry {
	return newEntry(name, cfg)
}
