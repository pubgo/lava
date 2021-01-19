package golug_grpc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/golug_base"
	registry "github.com/pubgo/golug/golug_registry"
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
	*golug_base.Entry
	mu                       sync.RWMutex
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
	if err := registerGw(fmt.Sprintf("localhost:%d", t.Options().Port), app.Group("/")); err != nil {
		return err
	}

	var data []map[string]string
	for i, stacks := range app.Stack() {
		data = append(data, make(map[string]string))
		for _, stack := range stacks {
			if stack == nil {
				continue
			}

			if stack.Path == "/" {
				continue
			}
			data[i][stack.Method] = stack.Path
		}
	}
	fmt.Printf("%#v\n", data)

	return app.Listen(t.cfg.GwAddr)
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

	ts := xerror.PanicErr(net.Listen("tcp", fmt.Sprintf(":%d", t.Options().Port))).(net.Listener)
	xlog.Infof("Server [grpc] Listening on %s", ts.Addr().String())

	cancel := xprocess.GoDelay(time.Second, func(ctx context.Context) {
		defer xerror.Resp(func(err xerror.XErr) { xlog.Error("grpcEntry.Start handle error", xlog.Any("err", err)) })

		if err := t.srv.Serve(ts); err != nil && err != grpc.ErrServerStopped {
			xlog.Error(err.Error())
		}
		return
	})

	t.WithBeforeStop(func(_ *golug_entry.BeforeStop) { cancel() })

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

func newEntry(name string) *grpcEntry {
	ent := &grpcEntry{Entry: golug_base.New(name)}
	ent.initFlags()

	// 服务启动后, 启动网关
	ent.WithAfterStart(func(_ *golug_entry.AfterStart) { xerror.Panic(ent.startGw()) })
	ent.OnCfgWithName(Name, &ent.cfg)
	return ent
}

func New(name string) Entry { return newEntry(name) }
