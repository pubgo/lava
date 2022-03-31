package service_builder

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	fiber2 "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/cmux"
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/logging/logutil"
	encoding3 "github.com/pubgo/lava/encoding"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/pkg/fiber_builder"
	"github.com/pubgo/lava/pkg/grpc_builder"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/version"
)

func New(name string, desc string, plugins ...plugin.Plugin) service.Service {
	return newService(name, desc, plugins...)
}

func newService(name string, desc string, plugins ...plugin.Plugin) *serviceImpl {
	var g = &serviceImpl{
		ctx:        context.Background(),
		pluginList: plugins,
		cmd:        &cli.Command{Name: name, Usage: desc},
		srv:        grpc_builder.New(),
		gw:         fiber_builder.New(),
		inproc:     &inprocgrpc.Channel{},
		app:        fiber2.New(),
		net:        cmux.DefaultCfg(),
		cfg: Cfg{
			name:     name,
			hostname: runtime.Hostname,
			id:       uuid.New().String(),
			Grpc:     grpc_builder.GetDefaultCfg(),
			Gw:       fiber_builder.Cfg{},
		},
	}

	return g
}

type serviceImpl struct {
	beforeStarts []func()
	afterStarts  []func()
	beforeStops  []func()
	afterStops   []func()
	pluginList   []plugin.Plugin
	middlewares  []service.Middleware
	services     []service.Desc

	cmd *cli.Command

	L *logging.Logger `name:"service"`

	net *cmux.Mux

	cfg Cfg
	srv grpc_builder.Builder
	gw  fiber_builder.Builder
	app *fiber2.App

	// inproc Channel is used to serve grpc gateway
	inproc *inprocgrpc.Channel

	wrapperUnary  service.HandlerFunc
	wrapperStream service.HandlerFunc

	ctx context.Context
}

func (t *serviceImpl) Ctx() context.Context { return t.ctx }

func (t *serviceImpl) RegisterRouter(prefix string, handlers ...fiber2.Handler) fiber2.Router {
	return t.app.Group(prefix, handlers...)
}

func (t *serviceImpl) RegisterApp(prefix string, r *fiber2.App) {
	t.app.Mount(prefix, r)
}

func (t *serviceImpl) Middleware(mid service.Middleware) {
	if mid == nil {
		return
	}

	t.middlewares = append(t.middlewares, mid)
}

func (t *serviceImpl) Middlewares() []service.Middleware { return t.middlewares }

func (t *serviceImpl) plugins() []plugin.Plugin { return t.pluginList }

func (t *serviceImpl) middleware(mid service.Middleware) {
	if mid == nil {
		return
	}

	t.middlewares = append(t.middlewares, mid)
}

func (t *serviceImpl) BeforeStarts(f ...func()) { t.beforeStarts = append(t.beforeStarts, f...) }
func (t *serviceImpl) BeforeStops(f ...func())  { t.beforeStops = append(t.beforeStops, f...) }
func (t *serviceImpl) AfterStarts(f ...func())  { t.afterStarts = append(t.afterStarts, f...) }
func (t *serviceImpl) AfterStops(f ...func())   { t.afterStops = append(t.afterStops, f...) }
func (t *serviceImpl) Plugin(plugin plugin.Plugin) {
	if plugin == nil {
		return
	}
	t.pluginList = append(t.pluginList, plugin)
}

func (t *serviceImpl) init() error {
	defer xerror.RespExit()

	t.net.Addr = runtime.Addr

	// 依赖对象注入
	inject.Inject(t)

	// 编码注册
	encoding3.Each(func(_ string, cdc encoding3.Codec) {
		encoding.RegisterCodec(cdc)
	})

	// 配置解析
	_ = config.Decode(Name, &t.cfg)

	t.initDebug()

	// 网关初始化
	xerror.Panic(t.gw.Build(t.cfg.Gw))
	t.gw.Get().Use(t.handlerHttpMiddle(t.middlewares))
	t.gw.Get().Mount("/", t.app)

	// 注册系统middleware
	t.srv.UnaryInterceptor(t.handlerUnaryMiddle(t.middlewares))
	t.srv.StreamInterceptor(t.handlerStreamMiddle(t.middlewares))

	// grpc serve初始化
	xerror.Panic(t.srv.Build(t.cfg.Grpc))

	// 加载inproc的middleware
	t.inproc.WithServerUnaryInterceptor(t.handlerUnaryMiddle(t.middlewares))
	t.inproc.WithServerStreamInterceptor(t.handlerStreamMiddle(t.middlewares))

	// 初始化 handlers
	for _, srv := range t.ServiceDesc() {
		// service handler依赖对象注入
		xerror.Panic(xerror.Try(func() {
			inject.Inject(srv.Handler)
			t.L.Info("Service Handler Injection", zap.String("handler", fmt.Sprintf("%#v", srv.Handler)))
		}))

		if h, ok := srv.Handler.(service.Handler); ok {
			t.AfterStops(h.Close)
			logutil.LogOrPanic(t.L, "Service Handler Init", func() error {
				return xerror.Try(func() {
					// register router
					h.Router(t.gw.Get())
					// service handler init
					h.Init()
				})
			})
		}
	}
	return nil
}

func (t *serviceImpl) Flags(flags ...cli.Flag) {
	if len(flags) == 0 {
		return
	}

	t.cmd.Flags = append(t.cmd.Flags, flags...)
}

func (t *serviceImpl) command() *cli.Command { return t.cmd }

func (t *serviceImpl) RegisterMatcher(priority int64, matches ...cmux.Matcher) chan net.Listener {
	return t.net.Register(priority, matches...)
}

func (t *serviceImpl) ServiceDesc() []service.Desc { return t.services }

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

func (t *serviceImpl) GrpcClientInnerConn() grpc.ClientConnInterface { return t.inproc }

func (t *serviceImpl) RegisterService(desc service.Desc) {
	xerror.Assert(desc.Handler == nil, "[handler] is nil")

	t.srv.RegisterService(&desc.ServiceDesc, desc.Handler)
	t.inproc.RegisterService(&desc.ServiceDesc, desc.Handler)
	t.services = append(t.services, desc)

	if h, ok := desc.Handler.(service.Handler); ok {
		t.Flags(h.Flags()...)
	}
}

func (t *serviceImpl) start() (gErr error) {
	defer xerror.RespErr(&gErr)

	logutil.OkOrPanic(t.L, "service before-start", func() error {
		var beforeList []func()
		for _, p := range plugin.All() {
			beforeList = append(beforeList, p.BeforeStarts()...)
		}
		beforeList = append(beforeList, t.beforeStarts...)
		for i := range beforeList {
			t.L.Sugar().Infof("running %s", stack.Func(beforeList[i]))
			xerror.PanicF(xerror.Try(beforeList[i]), stack.Func(beforeList[i]))
		}
		return nil
	})

	var grpcLn = t.net.HTTP2()
	var gwLn = t.net.HTTP1Fast()

	// 启动grpc网关
	syncx.GoDelay(func() {
		t.L.Info("[grpc-gw] Server Starting")
		logutil.LogOrErr(t.L, "[grpc-gw] Server Stop", func() error {
			if err := t.gw.Get().Listener(<-gwLn); err != nil &&
				!errors.Is(err, cmux.ErrListenerClosed) &&
				!errors.Is(err, http.ErrServerClosed) &&
				!errors.Is(err, net.ErrClosed) {
				return err
			}
			return nil
		})
	})

	logutil.OkOrPanic(t.L, "service start", func() error {
		t.L.Sugar().Infof("Server Listening on http://%s:%d", netutil.GetLocalIP(), netutil.MustGetPort(runtime.Addr))

		// 启动grpc服务
		syncx.GoDelay(func() {
			t.L.Info("[grpc] Server Starting")
			logutil.LogOrErr(t.L, "[grpc] Server Stop", func() error {
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
			t.L.Info("[cmux] Server Starting")
			logutil.LogOrErr(t.L, "[cmux] Server Stop", func() error {
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

	logutil.OkOrPanic(t.L, "service after-start", func() error {
		var afterList []func()
		for _, p := range plugin.All() {
			afterList = append(afterList, p.AfterStarts()...)
		}
		afterList = append(afterList, t.afterStarts...)
		for i := range afterList {
			t.L.Sugar().Infof("running %s", stack.Func(afterList[i]))
			xerror.PanicF(xerror.Try(afterList[i]), stack.Func(afterList[i]))
		}
		return nil
	})
	return nil
}

func (t *serviceImpl) stop() (err error) {
	defer xerror.RespErr(&err)

	logutil.OkOrErr(t.L, "service before-stop", func() error {
		var beforeList []func()
		for _, p := range plugin.All() {
			beforeList = append(beforeList, p.BeforeStops()...)
		}
		beforeList = append(beforeList, t.beforeStops...)
		for i := range beforeList {
			t.L.Sugar().Infof("running %s", stack.Func(beforeList[i]))
			xerror.PanicF(xerror.Try(beforeList[i]), stack.Func(beforeList[i]))
		}
		return nil
	})

	logutil.LogOrErr(t.L, "[grpc] GracefulStop", func() error {
		t.srv.Get().GracefulStop()
		xerror.Panic(t.gw.Get().Shutdown())
		xerror.Panic(t.net.Close())
		return nil
	})

	logutil.OkOrErr(t.L, "service after-stop", func() error {
		var afterList []func()
		for _, p := range plugin.All() {
			afterList = append(afterList, p.AfterStops()...)
		}
		afterList = append(afterList, t.afterStops...)
		for i := range afterList {
			t.L.Sugar().Infof("running %s", stack.Func(afterList[i]))
			xerror.PanicF(xerror.Try(afterList[i]), stack.Func(afterList[i]))
		}
		return nil
	})

	return
}
