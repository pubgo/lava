package golug_grpc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/golug_base"
	registry "github.com/pubgo/golug/golug_registry"
	"github.com/pubgo/golug/pkg/golug_utils"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess"
	"github.com/spf13/pflag"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"
)

var _ Entry = (*grpcEntry)(nil)

type grpcEntry struct {
	mu sync.RWMutex
	golug_entry.Entry
	cfg                      Cfg
	registry                 registry.Registry
	registryMap              map[string][]*registry.Endpoint
	registered               atomic.Bool
	handlers                 []interface{}
	endpoints                []*registry.Endpoint
	srv                      *grpc.Server
	opts                     []grpc.ServerOption
	unaryServerInterceptors  []grpc.UnaryServerInterceptor
	streamServerInterceptors []grpc.StreamServerInterceptor
}

// EnableDebug
// https://github.com/grpc/grpc-experiments/tree/master/gdebug
func (t *grpcEntry) EnableDebug() { service.RegisterChannelzServiceToServer(t.srv) }
func (t *grpcEntry) RegisterUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.unaryServerInterceptors = append(t.unaryServerInterceptors, interceptors...)
}

func (t *grpcEntry) RegisterStreamInterceptor(interceptors ...grpc.StreamServerInterceptor) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.streamServerInterceptors = append(t.streamServerInterceptors, interceptors...)
}

func (t *grpcEntry) Init() (err error)            { return xerror.Wrap(t.Entry.Run().Init()) }
func (t *grpcEntry) Options() golug_entry.Options { return t.Entry.Run().Options() }
func (t *grpcEntry) Run() golug_entry.RunEntry    { return t }
func (t *grpcEntry) UnWrap(fn interface{})        { xerror.Next().Panic(golug_utils.UnWrap(t, fn)) }
func (t *grpcEntry) Register(handler interface{}, opts ...Option) {
	xerror.Assert(handler == nil, "[handler] should not be nil")

	t.mu.Lock()
	defer t.mu.Unlock()
	t.handlers = append(t.handlers, handler)
}

// 开启api网关模式
func (t *grpcEntry) startGw() (err error) {
	if t.cfg.GwAddr == "" {
		return nil
	}

	app := fiber.New()

	// 开启api网关模式
	return registerGw(t.cfg.GwAddr, app.Group("/"))
}

func (t *grpcEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	// 初始化server
	t.srv = grpc.NewServer(append(t.opts,
		grpc.ChainUnaryInterceptor(t.unaryServerInterceptors...),
		grpc.ChainStreamInterceptor(t.streamServerInterceptors...))...,
	)

	t.endpoints = t.endpoints[:0]
	// 初始化routes
	for i := range t.handlers {
		t.endpoints = append(t.endpoints, newRpcHandler(t.handlers[i])...)
		xerror.Panic(register(t.srv, t.handlers[i]))
	}

	// 方便grpcurl调用和调试
	reflection.Register(t.srv)

	cancel := xprocess.GoDelay(time.Second, func(ctx context.Context) {
		defer xerror.Resp(func(err xerror.XErr) {
			xlog.Error("grpcEntry.Start handle error", xlog.Any("err", err))
		})

		//server_addr, err := net.ResolveUnixAddr("unix", server_file)
		//if err != nil {
		//	log.Fatal("fialed to resolve unix addr")
		//}
		//
		//lis, err := net.ListenUnix("unix", server_addr)
		//if err != nil {
		//	log.Fatal("failed to listen: %v", err)
		//}

		ts := xerror.PanicErr(net.Listen("tcp", fmt.Sprintf(":%d", t.Options().Port))).(net.Listener)
		xlog.Infof("Server [grpc] Listening on %s", ts.Addr().String())
		if err := t.srv.Serve(ts); err != nil && err != grpc.ErrServerStopped {
			xlog.Error(err.Error())
		}
		return
	})

	xerror.Panic(dix_run.WithBeforeStop(func(ctx *dix_run.BeforeStopCtx) { cancel() }))

	return nil
}

func (t *grpcEntry) Stop() (err error) {
	defer xerror.RespErr(&err)
	t.srv.GracefulStop()
	xlog.Infof("Server [grpc] Closed OK")
	return nil
}

func (t *grpcEntry) initFlags() {
	t.Flags(func(flags *pflag.FlagSet) {
		flags.StringVar(&t.cfg.GwAddr, "gw", t.cfg.GwAddr, "set addr and enable gateway mode")
	})
}

func newEntry(name string, cfg interface{}) *grpcEntry {
	ent := &grpcEntry{Entry: golug_base.New(name, cfg)}
	ent.initFlags()

	// 服务启动后, 启动网关
	xerror.Panic(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) { xerror.Panic(ent.startGw()) }))
	return ent
}

func New(name string, cfg interface{}) *grpcEntry {
	return newEntry(name, cfg)
}

func UnixConnect(addr string, t time.Duration) (net.Conn, error) {
	unix_addr, err := net.ResolveUnixAddr("unix", "")
	conn, err := net.DialUnix("unix", nil, unix_addr)
	return conn, err
}
