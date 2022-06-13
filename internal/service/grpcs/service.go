package grpcs

import (
	"errors"
	"fmt"
	"github.com/pubgo/lava/config"
	"net"
	"net/http"
	"strings"

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

	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/core/router"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/core/signal"
	cmux2 "github.com/pubgo/lava/internal/cmux"
	fiber_builder2 "github.com/pubgo/lava/internal/pkg/fiber_builder"
	grpc_builder2 "github.com/pubgo/lava/internal/pkg/grpc_builder"
	netutil2 "github.com/pubgo/lava/internal/pkg/netutil"
	"github.com/pubgo/lava/internal/pkg/syncx"
	"github.com/pubgo/lava/internal/pkg/utils"
	"github.com/pubgo/lava/logging/logutil"
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
			Grpc: grpc_builder2.GetDefaultCfg(),
			Api:  &fiber_builder2.Cfg{},
		},
		Lifecycle: lifecycle.New(),
		grpcSrv:   grpc_builder2.New(),
		httpSrv:   fiber_builder2.New(),
		handlers:  grpchan.HandlerMap{},
	}

	g.cmd.Before = func(context *cli.Context) (gErr error) {
		defer xerror.RecoverErr(&gErr, func(err xerror.XErr) xerror.XErr {
			fmt.Println(dix.Graph())
			return err
		})

		if runmode.Project == "" {
			runmode.Project = strings.Split(name, " ")[0]
		}
		xerror.Assert(runmode.Project == "", "project is null")
		return
	}

	g.cmd.Action = func(ctx *cli.Context) error {
		defer xerror.RecoverAndExit()
		xerror.Panic(g.start())
		signal.Block()
		xerror.Panic(g.stop())
		return nil
	}

	return g
}

var _ service.Service = (*serviceImpl)(nil)

type serviceImpl struct {
	lifecycle.Lifecycle
	Middlewares []service.Middleware
	Lc          lifecycle.Lifecycle
	Log         *zap.Logger
	Net         *cmux2.Mux
	App         *router.App
	Register    *registry.Loader

	cmd      *cli.Command
	cfg      Cfg
	grpcSrv  grpc_builder2.Builder
	httpSrv  fiber_builder2.Builder
	deps     []interface{}
	handlers grpchan.HandlerMap
}

func (t *serviceImpl) init() (gErr error) {
	defer xerror.RecoverErr(&gErr)
	t.Log = t.Log.Named(runmode.Project)

	dix.Inject(func(cfg config.Config) {
		// 配置解析
		xerror.Panic(cfg.UnmarshalKey(Name, &t.cfg))
	})

	for i := range t.deps {
		if t.deps[i] == nil {
			continue
		}

		dix.Inject(t.deps[i])
	}

	var middlewares []service.Middleware
	for _, m := range t.Middlewares {
		middlewares = append(middlewares, m)
	}

	unaryInt := t.handlerUnaryMiddle(middlewares)
	streamInt := t.handlerStreamMiddle(middlewares)

	httpgrpc.HandleServices(func(pattern string, handler func(http.ResponseWriter, *http.Request)) {
		t.App.Post(pattern, func(ctx *fiber2.Ctx) error {
			ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
			ctx.Response().Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			ctx.Response().Header.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, X-Extra-Header, Content-Type, Accept, Authorization")
			ctx.Response().Header.Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			return adaptor.HTTPHandlerFunc(handler)(ctx)
		})
	}, "/"+runmode.Project, t.handlers, unaryInt, streamInt)

	// 注册系统middleware
	t.grpcSrv.UnaryInterceptor(unaryInt)
	t.grpcSrv.StreamInterceptor(streamInt)

	// grpc serve初始化
	xerror.Panic(t.grpcSrv.Build(t.cfg.Grpc))

	// 初始化 handlers
	t.handlers.ForEach(func(desc *grpc.ServiceDesc, svr interface{}) {
		t.grpcSrv.Get().RegisterService(desc, svr)

		if h, ok := svr.(service.Close); ok {
			t.AfterStops(h.Close)
		}

		if h, ok := svr.(service.Init); ok {
			h.Init()
		}

		if h, ok := svr.(service.WebHandler); ok {
			h.Router(t.App)
		}
	})

	if t.cfg.PrintRoute {
		for _, stacks := range t.App.Stack() {
			for _, s := range stacks {
				t.Log.Info("service route",
					zap.String("name", s.Name),
					zap.String("path", s.Path),
					zap.String("method", s.Method),
				)
			}
		}
	}

	// 网关初始化
	xerror.Panic(t.httpSrv.Build(t.cfg.Api))
	t.httpSrv.Get().Use(t.handlerHttpMiddle(middlewares))
	t.httpSrv.Get().Mount("/", t.App.App)
	return nil
}

func (t *serviceImpl) SubCmd(cmd *cli.Command) {
	t.cmd.Subcommands = append(t.cmd.Subcommands, cmd)
}

func (t *serviceImpl) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	xerror.Assert(desc == nil, "[desc] is nil")
	xerror.Assert(impl == nil, "[impl] is nil")
	t.handlers.RegisterService(desc, impl)
}

func (t *serviceImpl) Provider(regs interface{}) {
	t.deps = append(t.deps, regs)
}

func (t *serviceImpl) Command() *cli.Command { return t.cmd }

func (t *serviceImpl) Options() service.Options {
	return service.Options{
		Name:      runmode.Project,
		Id:        runmode.InstanceID,
		Version:   version.Version,
		Port:      netutil2.MustGetPort(t.Net.Addr),
		Addr:      t.Net.Addr,
		Advertise: "",
	}
}

func (t *serviceImpl) start() (gErr error) {
	defer xerror.RecoverErr(&gErr)

	xerror.Panic(t.init())

	logutil.OkOrPanic(t.Log, "service before-start", func() error {
		for _, run := range append(t.Lc.GetBeforeStarts(), t.GetBeforeStarts()...) {
			t.Log.Sugar().Infof("before-start running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})

	var grpcLn = t.Net.HTTP2()
	var gwLn = t.Net.HTTP1()

	logutil.OkOrPanic(t.Log, "service start", func() error {
		t.Log.Sugar().Infof("Server Listening on http://%s:%d", netutil2.GetLocalIP(), netutil2.MustGetPort(t.Net.Addr))

		// 启动grpc网关
		syncx.GoDelay(func() {
			t.Log.Info("[grpc-gw] Server Starting")
			logutil.LogOrErr(t.Log, "[grpc-gw] Server Stop", func() error {
				if err := t.httpSrv.Get().Listener(<-gwLn); err != nil &&
					!errors.Is(err, cmux2.ErrListenerClosed) &&
					!errors.Is(err, http.ErrServerClosed) &&
					!errors.Is(err, net.ErrClosed) {
					return err
				}
				return nil
			})
		})

		// 启动grpc服务
		syncx.GoDelay(func() {
			t.Log.Info("[grpc] Server Starting")
			logutil.LogOrErr(t.Log, "[grpc] Server Stop", func() error {
				if err := t.grpcSrv.Get().Serve(<-grpcLn); err != nil &&
					err != cmux2.ErrListenerClosed &&
					!errors.Is(err, http.ErrServerClosed) &&
					!errors.Is(err, net.ErrClosed) {
					return err
				}
				return nil
			})
		})

		// 启动net网络
		syncx.GoDelay(func() {
			t.Log.Info("[cmux] Server Starting")
			logutil.LogOrErr(t.Log, "[cmux] Server Stop", func() error {
				if err := t.Net.Serve(); err != nil &&
					!errors.Is(err, http.ErrServerClosed) &&
					!errors.Is(err, net.ErrClosed) {
					return err
				}
				return nil
			})
		})
		return nil
	})

	logutil.OkOrPanic(t.Log, "service after-start", func() error {
		for _, run := range append(t.Lc.GetAfterStarts(), t.GetAfterStarts()...) {
			t.Log.Sugar().Infof("after-start running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})
	return nil
}

func (t *serviceImpl) stop() (err error) {
	defer xerror.RecoverErr(&err)

	logutil.OkOrErr(t.Log, "service before-stop", func() error {
		for _, run := range append(t.Lc.GetBeforeStops(), t.GetBeforeStops()...) {
			t.Log.Sugar().Infof("before-stop running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})

	logutil.LogOrErr(t.Log, "[grpc-gw] Shutdown", func() error {
		xerror.Panic(t.httpSrv.Get().Shutdown())
		return nil
	})

	logutil.LogOrErr(t.Log, "[grpc] GracefulStop", func() error {
		t.grpcSrv.Get().GracefulStop()
		return t.Net.Close()
	})

	logutil.OkOrErr(t.Log, "service after-stop", func() error {
		for _, run := range append(t.Lc.GetAfterStops(), t.GetAfterStops()...) {
			t.Log.Sugar().Infof("after-stop running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})

	return
}
