package service_builder

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"plugin"

	"github.com/fullstorydev/grpchan/inprocgrpc"
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
	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/internal/envs"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/middleware"
	"github.com/pubgo/lava/module"
	"github.com/pubgo/lava/pkg/fiber_builder"
	"github.com/pubgo/lava/pkg/grpc_builder"
	"github.com/pubgo/lava/pkg/gw_builder"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/version"
)

func New(name string, desc string, plugins ...plugin.Plugin) service.Service {
	return newService(name, desc, plugins...)
}

func newService(name string, desc string, plugins ...plugin.Plugin) *serviceImpl {
	var g = &serviceImpl{
		log:        zap.L().Named(name),
		ctx:        context.Background(),
		pluginList: plugins,
		cmd: &cli.Command{
			Name:  name,
			Usage: desc,
			Flags: flags.GetFlags(),
		},
		srv:     grpc_builder.New(),
		gw:      fiber_builder.New(),
		inproc:  &inprocgrpc.Channel{},
		httpSrv: fiber2.New(),
		net:     cmux.DefaultCfg(),
		cfg: Cfg{
			name:     name,
			hostname: runtime.Hostname,
			id:       runtime.AppID,
			Grpc:     grpc_builder.GetDefaultCfg(),
			Gw:       fiber_builder.Cfg{},
		},
	}

	g.cmd.Action = func(ctx *cli.Context) error {
		defer xerror.RespExit()
		xerror.Panic(g.init())
		xerror.Panic(g.start())
		signal.Block()
		xerror.Panic(g.stop())
		return nil
	}

	g.Provide(func() service.Service { return g })
	return g
}

var _ service.Service = (*serviceImpl)(nil)

type serviceImpl struct {
	beforeStarts []func()
	afterStarts  []func()
	beforeStops  []func()
	afterStops   []func()
	pluginList   []plugin.Plugin
	middlewares  []middleware.Middleware
	services     []service.Desc

	log *zap.Logger
	app *fx.App
	cmd *cli.Command

	net *cmux.Mux

	cfg     Cfg
	srv     grpc_builder.Builder
	gw      fiber_builder.Builder
	httpSrv *fiber2.App
	opts    []fx.Option

	// inproc Channel is used to serve grpc gateway
	inproc *inprocgrpc.Channel

	wrapperUnary  middleware.HandlerFunc
	wrapperStream middleware.HandlerFunc

	ctx        context.Context
	gwHandlers []func(ctx context.Context, mux *gw.ServeMux, cc grpc.ClientConnInterface) error
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

	t.srv.RegisterService(&desc.ServiceDesc, desc.Handler)
	t.inproc.RegisterService(&desc.ServiceDesc, desc.Handler)
	t.services = append(t.services, desc)

	if h, ok := desc.Handler.(abc.Flags); ok {
		t.Flags(h.Flags()...)
	}

	t.opts = append(t.opts, fx.Populate(desc.Handler))
}

func (t *serviceImpl) RegRouter(prefix string, fn func(r fiber2.Router)) {
	t.httpSrv.Route(prefix, fn)
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

	// 项目名初始化
	runtime.Project = t.cmd.Name
	envs.SetName(version.Domain, runtime.Project)

	// 运行环境检查
	if _, ok := runtime.RunModeValue[runtime.Mode.String()]; !ok {
		panic(fmt.Sprintf("mode(%s) not match in (%v)", runtime.Mode, runtime.RunModeValue))
	}

	t.app = fx.New(append(module.List(), t.opts...)...)

	t.net.Addr = runtime.Addr

	// 配置解析
	xerror.Panic(config.UnmarshalKey(Name, &t.cfg))

	// 网关初始化
	xerror.Panic(t.gw.Build(t.cfg.Gw))
	t.gw.Get().Use(t.handlerHttpMiddle(t.middlewares))
	t.gw.Get().Mount("/", t.httpSrv)

	// 注册系统middleware
	t.srv.UnaryInterceptor(t.handlerUnaryMiddle(t.middlewares))
	t.srv.StreamInterceptor(t.handlerStreamMiddle(t.middlewares))

	// grpc serve初始化
	xerror.Panic(t.srv.Build(t.cfg.Grpc))

	// 加载inproc的middleware
	t.inproc.WithServerUnaryInterceptor(t.handlerUnaryMiddle(t.middlewares))
	t.inproc.WithServerStreamInterceptor(t.handlerStreamMiddle(t.middlewares))

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

	var cfg = gw_builder.DefaultCfg()
	xerror.Panic(config.UnmarshalKey(Name, &cfg))

	var builder = gw_builder.New()
	xerror.Panic(builder.Build(cfg))
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
		Name:      t.cfg.name,
		Id:        t.cfg.id,
		Version:   version.Version,
		Port:      netutil.MustGetPort(t.net.Addr),
		Address:   t.net.Addr,
		Advertise: t.cfg.Advertise,
	}
}

func (t *serviceImpl) start() (gErr error) {
	defer xerror.RespErr(&gErr)

	logutil.OkOrPanic(t.log, "service before-start", func() error {
		for i := range t.beforeStarts {
			t.log.Sugar().Infof("running %s", stack.Func(t.beforeStarts[i]))
			xerror.PanicF(xerror.Try(t.beforeStarts[i]), stack.Func(t.beforeStarts[i]))
		}
		return nil
	})

	var grpcLn = t.net.HTTP2()
	var gwLn = t.net.HTTP1Fast()

	// 启动grpc网关
	syncx.GoDelay(func() {
		t.log.Info("[grpc-gw] Server Starting")
		logutil.LogOrErr(t.log, "[grpc-gw] Server Stop", func() error {
			if err := t.gw.Get().Listener(<-gwLn); err != nil &&
				!errors.Is(err, cmux.ErrListenerClosed) &&
				!errors.Is(err, http.ErrServerClosed) &&
				!errors.Is(err, net.ErrClosed) {
				return err
			}
			return nil
		})
	})

	logutil.OkOrPanic(t.log, "service start", func() error {
		t.log.Sugar().Infof("Server Listening on http://%s:%d", netutil.GetLocalIP(), netutil.MustGetPort(runtime.Addr))

		// 启动grpc服务
		syncx.GoDelay(func() {
			t.log.Info("[grpc] Server Starting")
			logutil.LogOrErr(t.log, "[grpc] Server Stop", func() error {
				if err := t.srv.Get().Serve(<-grpcLn); err != nil &&
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
		for i := range t.afterStarts {
			t.log.Sugar().Infof("running %s", stack.Func(t.afterStarts[i]))
			xerror.PanicF(xerror.Try(t.afterStarts[i]), stack.Func(t.afterStarts[i]))
		}
		return nil
	})
	return nil
}

func (t *serviceImpl) stop() (err error) {
	defer xerror.RespErr(&err)

	logutil.OkOrErr(t.log, "service before-stop", func() error {
		for i := range t.beforeStops {
			t.log.Sugar().Infof("running %s", stack.Func(t.beforeStops[i]))
			xerror.PanicF(xerror.Try(t.beforeStops[i]), stack.Func(t.beforeStops[i]))
		}
		return nil
	})

	logutil.LogOrErr(t.log, "[grpc] GracefulStop", func() error {
		t.srv.Get().GracefulStop()
		xerror.Panic(t.gw.Get().Shutdown())
		xerror.Panic(t.net.Close())
		return nil
	})

	logutil.OkOrErr(t.log, "service after-stop", func() error {
		for i := range t.afterStops {
			t.log.Sugar().Infof("running %s", stack.Func(t.afterStops[i]))
			xerror.PanicF(xerror.Try(t.afterStops[i]), stack.Func(t.afterStops[i]))
		}
		return nil
	})

	return
}
