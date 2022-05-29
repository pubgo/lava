package service_builder

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/fullstorydev/grpchan"
	"github.com/fullstorydev/grpchan/httpgrpc"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/gofiber/adaptor/v2"
	fiber2 "github.com/gofiber/fiber/v2"
	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/abc"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/cmux"
	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/middleware"
	"github.com/pubgo/lava/pkg/fiber_builder"
	"github.com/pubgo/lava/pkg/grpc_builder"
	"github.com/pubgo/lava/pkg/gw_builder"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/pkg/utils"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/version"
)

func New(name string, desc ...string) service.Service {
	return newService(name, desc...)
}

func newService(name string, desc ...string) *serviceImpl {
	var g = &serviceImpl{
		cmd: &cli.Command{
			Name:  name,
			Usage: utils.FirstNotEmpty(append(desc, fmt.Sprintf("%s service", name))...),
			Flags: flags.GetFlags(),
		},

		cfg: Cfg{
			Grpc: grpc_builder.GetDefaultCfg(),
			Api:  fiber_builder.Cfg{},
			Gw:   gw_builder.DefaultCfg(),
		},
		log:     zap.L().Named(runtime.Project),
		grpcSrv: grpc_builder.New(),
		api:     fiber_builder.New(),
		inproc:  &inprocgrpc.Channel{},
		httpSrv: fiber2.New(),
		net:     cmux.DefaultCfg(),
	}

	g.cmd.Action = func(ctx *cli.Context) error {
		defer xerror.RespExit()
		xerror.Panic(g.start())
		signal.Block()
		xerror.Panic(g.stop())
		return nil
	}

	g.Provide(func() service.App { return g })
	g.Invoke(func(m lifecycle.GetLifecycle) { g.modules = m })
	return g
}

var _ service.Service = (*serviceImpl)(nil)

type serviceImpl struct {
	beforeStarts []func()
	afterStarts  []func()
	beforeStops  []func()
	afterStops   []func()
	middlewares  []middleware.Middleware
	services     []service.Desc

	modules lifecycle.GetLifecycle

	log *zap.Logger
	cmd *cli.Command

	net *cmux.Mux

	cfg     Cfg
	grpcSrv grpc_builder.Builder
	api     fiber_builder.Builder
	httpSrv *fiber2.App
	opts    []fx.Option

	// inproc Channel is used to serve grpc gateway
	inproc *inprocgrpc.Channel

	reg grpchan.HandlerMap

	wrapperUnary  middleware.HandlerFunc
	wrapperStream middleware.HandlerFunc

	ctx        context.Context
	gwHandlers []func(ctx context.Context, mux *gw.ServeMux, cc grpc.ClientConnInterface) error
}

func (t *serviceImpl) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	//TODO implement me
	panic("implement me")
}

func (t *serviceImpl) Provide(constructors ...interface{}) {
	t.opts = append(t.opts, fx.Provide(constructors...))
}

func (t *serviceImpl) Invoke(funcs ...interface{}) {
	t.opts = append(t.opts, fx.Invoke(funcs...))
}

func (t *serviceImpl) Start() error          { return t.start() }
func (t *serviceImpl) Stop() error           { return t.stop() }
func (t *serviceImpl) Command() *cli.Command { return t.cmd }

func (t *serviceImpl) RegService(desc service.Desc) {
	xerror.Assert(desc.Handler == nil, "[handler] is nil")

	t.grpcSrv.RegisterService(&desc.ServiceDesc, desc.Handler)
	t.inproc.RegisterService(&desc.ServiceDesc, desc.Handler)
	t.services = append(t.services, desc)

	t.opts = append(t.opts, fx.Populate(desc.Handler))

	//	s.reg.RegisterService(sd, ss)
}

func (t *serviceImpl) RegGateway(fn func(ctx context.Context, mux *gw.ServeMux, cc grpc.ClientConnInterface) error) {
	t.gwHandlers = append(t.gwHandlers, fn)
}

func (t *serviceImpl) RegApp(prefix string, r *fiber2.App) {
	t.httpSrv.Mount(prefix, r)
}

func (t *serviceImpl) Middleware(mid middleware.Middleware) {
	xerror.Assert(mid == nil, "[mid] is nil")
	t.middlewares = append(t.middlewares, mid)
}

func (t *serviceImpl) BeforeStarts(f ...func()) { t.beforeStarts = append(t.beforeStarts, f...) }
func (t *serviceImpl) BeforeStops(f ...func())  { t.beforeStops = append(t.beforeStops, f...) }
func (t *serviceImpl) AfterStarts(f ...func())  { t.afterStarts = append(t.afterStarts, f...) }
func (t *serviceImpl) AfterStops(f ...func())   { t.afterStops = append(t.afterStops, f...) }

func (t *serviceImpl) init() error {
	defer xerror.RespExit()

	for i := range t.opts {
		inject.Register(t.opts[i])
	}

	inject.Load()

	// 配置解析
	xerror.Panic(config.UnmarshalKey(Name, &t.cfg))

	t.net.Addr = runtime.Addr

	middlewares := t.middlewares[:]
	for _, m := range t.cfg.Middlewares {
		middlewares = append(middlewares, middleware.Get(m))
	}

	// 注册系统middleware
	t.grpcSrv.UnaryInterceptor(t.handlerUnaryMiddle(middlewares))
	t.grpcSrv.StreamInterceptor(t.handlerStreamMiddle(middlewares))

	// grpc serve初始化
	xerror.Panic(t.grpcSrv.Build(t.cfg.Grpc))

	// 加载inproc的middleware
	t.inproc.WithServerUnaryInterceptor(t.handlerUnaryMiddle(middlewares))
	t.inproc.WithServerStreamInterceptor(t.handlerStreamMiddle(middlewares))

	// 初始化 handlers
	for _, desc := range t.services {
		if h, ok := desc.Handler.(abc.Close); ok {
			t.AfterStops(h.Close)
		}

		if h, ok := desc.Handler.(abc.Init); ok {
			h.Init()
		}

		if h, ok := desc.Handler.(service.Handler); ok {
			h.Router(t.httpSrv)
		}
	}

	// gw builder
	var builder = gw_builder.New()
	xerror.Panic(builder.Build(t.cfg.Gw))
	var mux = builder.Get()

	for _, h := range t.gwHandlers {
		xerror.Panic(h(context.Background(), mux, t.inproc))
	}
	t.httpSrv.All(fmt.Sprintf("/api/%s/*", runtime.Project), adaptor.HTTPHandler(mux))

	if t.cfg.PrintRoute {
		for _, stacks := range t.httpSrv.Stack() {
			for _, s := range stacks {
				t.log.Info("service route",
					zap.String("name", s.Name),
					zap.String("path", s.Path),
					zap.String("method", s.Method),
				)
			}
		}
	}

	// 网关初始化
	xerror.Panic(t.api.Build(t.cfg.Api))
	t.api.Get().Use(t.handlerHttpMiddle(middlewares))
	t.api.Get().Mount("/", t.httpSrv)

	httpgrpc.HandleServices(func(pattern string, handler func(http.ResponseWriter, *http.Request)) {
		t.httpSrv.Post(pattern, func(ctx *fiber2.Ctx) error {
			ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
			ctx.Response().Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			ctx.Response().Header.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, X-Extra-Header, Content-Type, Accept, Authorization")
			ctx.Response().Header.Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			return adaptor.HTTPHandlerFunc(handler)(ctx)
		})
	}, "/", t.reg, nil, nil)

	return nil
}

func (t *serviceImpl) Flags(flags ...cli.Flag) {
	if len(flags) == 0 {
		return
	}

	t.cmd.Flags = append(t.cmd.Flags, flags...)
}

func (t *serviceImpl) Options() service.Options {
	return service.Options{
		Name:      runtime.Project,
		Id:        runtime.AppID,
		Version:   version.Version,
		Port:      netutil.MustGetPort(t.net.Addr),
		Address:   t.net.Addr,
		Advertise: t.cfg.Advertise,
	}
}

func (t *serviceImpl) start() (gErr error) {
	defer xerror.RespErr(&gErr)

	xerror.Panic(t.init())

	logutil.OkOrPanic(t.log, "service before-start", func() error {
		for _, run := range append(t.modules.GetBeforeStarts(), t.beforeStarts...) {
			t.log.Sugar().Infof("before-start running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})

	var grpcLn = t.net.HTTP2()
	var gwLn = t.net.HTTP1()

	logutil.OkOrPanic(t.log, "service start", func() error {
		t.log.Sugar().Infof("Server Listening on http://%s:%d", netutil.GetLocalIP(), netutil.MustGetPort(runtime.Addr))

		// 启动grpc网关
		syncx.GoDelay(func() {
			t.log.Info("[grpc-gw] Server Starting")
			logutil.LogOrErr(t.log, "[grpc-gw] Server Stop", func() error {
				if err := t.api.Get().Listener(<-gwLn); err != nil &&
					!errors.Is(err, cmux.ErrListenerClosed) &&
					!errors.Is(err, http.ErrServerClosed) &&
					!errors.Is(err, net.ErrClosed) {
					return err
				}
				return nil
			})
		})

		// 启动grpc服务
		syncx.GoDelay(func() {
			t.log.Info("[grpc] Server Starting")
			logutil.LogOrErr(t.log, "[grpc] Server Stop", func() error {
				if err := t.grpcSrv.Get().Serve(<-grpcLn); err != nil &&
					err != cmux.ErrListenerClosed &&
					!errors.Is(err, http.ErrServerClosed) &&
					!errors.Is(err, net.ErrClosed) {
					return err
				}
				return nil
			})
		})

		// 启动net网络
		syncx.GoDelay(func() {
			t.log.Info("[cmux] Server Starting")
			logutil.LogOrErr(t.log, "[cmux] Server Stop", func() error {
				if err := t.net.Serve(); err != nil &&
					!errors.Is(err, http.ErrServerClosed) &&
					!errors.Is(err, net.ErrClosed) {
					return err
				}
				return nil
			})
		})
		return nil
	})

	logutil.OkOrPanic(t.log, "service after-start", func() error {
		for _, run := range append(t.modules.GetAfterStarts(), t.afterStarts...) {
			t.log.Sugar().Infof("after-start running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})
	return nil
}

func (t *serviceImpl) stop() (err error) {
	defer xerror.RespErr(&err)

	logutil.OkOrErr(t.log, "service before-stop", func() error {
		for _, run := range append(t.modules.GetBeforeStops(), t.beforeStops...) {
			t.log.Sugar().Infof("before-stop running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})

	logutil.LogOrErr(t.log, "[grpc-gw] Shutdown", func() error {
		xerror.Panic(t.api.Get().Shutdown())
		return nil
	})

	logutil.LogOrErr(t.log, "[grpc] GracefulStop", func() error {
		t.grpcSrv.Get().GracefulStop()
		return t.net.Close()
	})

	logutil.OkOrErr(t.log, "service after-stop", func() error {
		for _, run := range append(t.modules.GetAfterStops(), t.afterStops...) {
			t.log.Sugar().Infof("after-stop running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})

	return
}
