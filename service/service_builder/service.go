package service_builder

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/fullstorydev/grpchan"
	"github.com/fullstorydev/grpchan/httpgrpc"
	"github.com/gofiber/adaptor/v2"
	fiber2 "github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/abc"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/cmux"
	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/middleware"
	"github.com/pubgo/lava/pkg/fiber_builder"
	"github.com/pubgo/lava/pkg/grpc_builder"
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
		Lifecycle: lifecycle.New(),
		cmd: &cli.Command{
			Name:  name,
			Usage: utils.FirstNotEmpty(append(desc, fmt.Sprintf("%s service", name))...),
			Flags: flags.GetFlags(),
		},
		cfg: Cfg{
			Grpc: grpc_builder.GetDefaultCfg(),
			Api:  fiber_builder.Cfg{},
		},
		grpcSrv:  grpc_builder.New(),
		api:      fiber_builder.New(),
		handlers: grpchan.HandlerMap{},
		httpSrv:  fiber2.New(),
		net:      cmux.DefaultCfg(),
	}

	g.cmd.Action = func(ctx *cli.Context) error {
		defer xerror.RespExit()
		xerror.Panic(g.start())
		signal.Block()
		xerror.Panic(g.stop())
		return nil
	}

	g.Dix(func() grpc.ServiceRegistrar { return g })
	g.Dix(func() service.App { return g })
	g.Dix(func(m lifecycle.Lifecycle) { g.lifecycle = m })
	g.Dix(func(log *logging.Logger) { g.log = log.Named(runtime.Project) })
	return g
}

var _ service.Service = (*serviceImpl)(nil)

type serviceImpl struct {
	lifecycle.Lifecycle
	middlewares []middleware.Middleware

	lifecycle lifecycle.Lifecycle

	log *zap.Logger
	cmd *cli.Command

	net *cmux.Mux

	cfg     Cfg
	grpcSrv grpc_builder.Builder
	api     fiber_builder.Builder
	httpSrv *fiber2.App
	opts    []interface{}

	handlers grpchan.HandlerMap

	wrapperUnary  middleware.HandlerFunc
	wrapperStream middleware.HandlerFunc

	ctx context.Context
}

func (t *serviceImpl) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	xerror.Assert(impl == nil, "[handler] is nil")
	xerror.Assert(desc == nil, "[desc] is nil")
	t.handlers.RegisterService(desc, impl)
}

func (t *serviceImpl) Dix(regs ...interface{}) {
	t.opts = append(t.opts, regs...)
}

func (t *serviceImpl) Start() error          { return t.start() }
func (t *serviceImpl) Stop() error           { return t.stop() }
func (t *serviceImpl) Command() *cli.Command { return t.cmd }

func (t *serviceImpl) RegApp(prefix string, r *fiber2.App) {
	t.httpSrv.Mount(prefix, r)
}

func (t *serviceImpl) Middleware(mid middleware.Middleware) {
	xerror.Assert(mid == nil, "[mid] is nil")
	t.middlewares = append(t.middlewares, mid)
}

func (t *serviceImpl) init() error {
	defer xerror.RespExit()

	for i := range t.opts {
		dix.Register(t.opts[i])
	}

	dix.Invoke()

	for k, g := range dix.Graph() {
		fmt.Println(k, g)
	}

	// 配置解析
	xerror.Panic(config.UnmarshalKey(Name, &t.cfg))

	t.net.Addr = runtime.Addr

	middlewares := t.middlewares[:]
	for _, m := range t.cfg.Middlewares {
		middlewares = append(middlewares, middleware.Get(m))
	}

	unaryInt := t.handlerUnaryMiddle(middlewares)
	streamInt := t.handlerStreamMiddle(middlewares)

	httpgrpc.HandleServices(func(pattern string, handler func(http.ResponseWriter, *http.Request)) {
		t.httpSrv.Post(pattern, func(ctx *fiber2.Ctx) error {
			ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
			ctx.Response().Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			ctx.Response().Header.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, X-Extra-Header, Content-Type, Accept, Authorization")
			ctx.Response().Header.Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			return adaptor.HTTPHandlerFunc(handler)(ctx)
		})
	}, "/"+runtime.Project, t.handlers, unaryInt, streamInt)

	// 注册系统middleware
	t.grpcSrv.UnaryInterceptor(unaryInt)
	t.grpcSrv.StreamInterceptor(streamInt)

	// grpc serve初始化
	xerror.Panic(t.grpcSrv.Build(t.cfg.Grpc))

	// 初始化 handlers
	t.handlers.ForEach(func(desc *grpc.ServiceDesc, svr interface{}) {
		t.grpcSrv.Get().RegisterService(desc, svr)

		if h, ok := svr.(abc.Close); ok {
			t.AfterStops(h.Close)
		}

		if h, ok := svr.(abc.Init); ok {
			h.Init()
		}

		if h, ok := svr.(service.WebHandler); ok {
			h.Router(t.httpSrv)
		}
	})

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
		for _, run := range append(t.lifecycle.GetBeforeStarts(), t.GetBeforeStarts()...) {
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
		for _, run := range append(t.lifecycle.GetAfterStarts(), t.GetAfterStarts()...) {
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
		for _, run := range append(t.lifecycle.GetBeforeStops(), t.GetBeforeStops()...) {
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
		for _, run := range append(t.lifecycle.GetAfterStops(), t.GetAfterStops()...) {
			t.log.Sugar().Infof("after-stop running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})

	return
}
