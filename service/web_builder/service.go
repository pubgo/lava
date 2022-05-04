package web_builder

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	fiber2 "github.com/gofiber/fiber/v2"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/pubgo/lava/abc"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/internal/running"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/middleware"
	"github.com/pubgo/lava/pkg/fiber_builder"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/pkg/utils"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/version"
)

func New(name string, desc ...string) service.Web {
	return newService(name, desc...)
}

func newService(name string, desc ...string) *webImpl {
	var g = &webImpl{
		cmd: &cli.Command{
			Name:  name,
			Usage: utils.FirstNotEmpty(append(desc, fmt.Sprintf("%s service", name))...),
			Flags: flags.GetFlags(),
		},
		cfg: Cfg{
			Api: fiber_builder.Cfg{},
		},
		log:     zap.L().Named(runtime.Project),
		api:     fiber_builder.New(),
		httpSrv: fiber2.New(),
	}

	g.cmd.Action = func(ctx *cli.Context) error {
		defer xerror.RespExit()
		xerror.Panic(g.start())
		signal.Block()
		xerror.Panic(g.stop())
		return nil
	}

	// 配置解析
	xerror.Panic(config.UnmarshalKey(Name, &g.cfg))

	g.Provide(func() service.App { return g })
	g.Invoke(func(m running.GetRunning) { g.modules = m })
	return g
}

var _ service.Web = (*webImpl)(nil)

type webImpl struct {
	beforeStarts []func()
	afterStarts  []func()
	beforeStops  []func()
	afterStops   []func()
	middlewares  []middleware.Middleware

	handlers []interface{}

	modules running.GetRunning

	log *zap.Logger
	cmd *cli.Command

	cfg     Cfg
	api     fiber_builder.Builder
	httpSrv *fiber2.App
	opts    []fx.Option
}

func (t *webImpl) Provide(constructors ...interface{}) {
	t.opts = append(t.opts, fx.Provide(constructors...))
}

func (t *webImpl) Invoke(funcs ...interface{}) {
	t.opts = append(t.opts, fx.Invoke(funcs...))
}

func (t *webImpl) Command() *cli.Command { return t.cmd }

func (t *webImpl) RegHandler(handler interface{}) {
	t.handlers = append(t.handlers, handler)
}

func (t *webImpl) RegApp(prefix string, r *fiber2.App) {
	t.httpSrv.Mount(prefix, r)
}

func (t *webImpl) Middleware(mid middleware.Middleware) {
	xerror.Assert(mid == nil, "[mid] is nil")
	t.middlewares = append(t.middlewares, mid)
}

func (t *webImpl) BeforeStarts(f ...func()) { t.beforeStarts = append(t.beforeStarts, f...) }
func (t *webImpl) BeforeStops(f ...func())  { t.beforeStops = append(t.beforeStops, f...) }
func (t *webImpl) AfterStarts(f ...func())  { t.afterStarts = append(t.afterStarts, f...) }
func (t *webImpl) AfterStops(f ...func())   { t.afterStops = append(t.afterStops, f...) }

func (t *webImpl) init() error {
	defer xerror.RespExit()

	inject.Init(append(inject.List(), t.opts...)...)

	middlewares := t.middlewares[:]
	for _, m := range t.cfg.Middlewares {
		middlewares = append(middlewares, middleware.Get(m))
	}

	for _, handler := range t.handlers {
		if h, ok := handler.(abc.Close); ok {
			t.BeforeStops(h.Close)
		}

		if h, ok := handler.(abc.Init); ok {
			h.Init()
		}

		if h, ok := handler.(service.Handler); ok {
			h.Router(t.httpSrv)
		}
	}

	if t.cfg.PrintRoute {
		for _, stacks := range t.httpSrv.Stack() {
			for _, s := range stacks {
				t.log.Info("service routes",
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

func (t *webImpl) Options() service.Options {
	return service.Options{
		Name:      runtime.Project,
		Id:        runtime.AppID,
		Version:   version.Version,
		Port:      netutil.MustGetPort(runtime.Addr),
		Address:   runtime.Addr,
		Advertise: t.cfg.Advertise,
	}
}

func (t *webImpl) Flags(flags ...cli.Flag) {
	if len(flags) == 0 {
		return
	}

	t.cmd.Flags = append(t.cmd.Flags, flags...)
}

func (t *webImpl) start() (gErr error) {
	defer xerror.RespErr(&gErr)

	xerror.Panic(t.init())

	logutil.OkOrPanic(t.log, "service before-start", func() error {
		for _, run := range append(t.modules.GetBeforeStarts(), t.beforeStarts...) {
			t.log.Sugar().Infof("before-start running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})

	logutil.OkOrPanic(t.log, "service start", func() error {
		t.log.Sugar().Infof("Server Listening on http://%s:%d", netutil.GetLocalIP(), netutil.MustGetPort(runtime.Addr))

		// 启动net网络
		syncx.GoDelay(func() {
			t.log.Info("[http] Server Starting")
			logutil.LogOrErr(t.log, "[http] Server Stop", func() error {
				if err := t.api.Get().Listen(runtime.Addr); err != nil &&
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

func (t *webImpl) stop() (err error) {
	defer xerror.RespErr(&err)

	logutil.OkOrErr(t.log, "service before-stop", func() error {
		for _, run := range append(t.modules.GetBeforeStops(), t.beforeStops...) {
			t.log.Sugar().Infof("before-stop running %s", stack.Func(run))
			logutil.ErrTry(t.log, run)
		}
		return nil
	})

	logutil.LogOrErr(t.log, "[http] Shutdown", func() error {
		logutil.ErrRecord(t.log, t.api.Get().Shutdown())
		return nil
	})

	logutil.OkOrErr(t.log, "service after-stop", func() error {
		for _, run := range append(t.modules.GetAfterStops(), t.afterStops...) {
			t.log.Sugar().Infof("after-stop running %s", stack.Func(run))
			logutil.ErrTry(t.log, run)
		}
		return nil
	})

	return
}
