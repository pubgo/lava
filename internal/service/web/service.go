package grpcs

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

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
	"github.com/pubgo/lava/core/runmode"
	cmux2 "github.com/pubgo/lava/internal/cmux"
	fiber_builder2 "github.com/pubgo/lava/internal/pkg/fiber_builder"
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
	var lc = lifecycle.New()
	var g *serviceImpl
	g = &serviceImpl{
		Lifecycle: lc,
		cmd: &cli.Command{
			Name:  name,
			Usage: utils.FirstNotEmpty(append(desc, fmt.Sprintf("%s service", name))...),
			Flags: flags.GetFlags(),
			Before: func(context *cli.Context) (gErr error) {
				defer xerror.RecoverErr(&gErr, func(err xerror.XErr) xerror.XErr {
					fmt.Println(dix.Graph())
					return err
				})

				if runmode.Project == "" {
					runmode.Project = strings.Split(name, " ")[0]
				}
				xerror.Assert(runmode.Project == "", "project is null")

				for i := range g.providerList {
					dix.Provider(g.providerList[i])
				}
				return
			},
		},
		lc:      lc,
		httpSrv: fiber_builder2.New(),
	}

	return g
}

var _ service.Service = (*serviceImpl)(nil)

type serviceImpl struct {
	lifecycle.Lifecycle
	Middlewares  []service.Middleware
	GetLifecycle lifecycle.GetLifecycle
	Log          *zap.Logger
	Net          *cmux2.Mux
	Routers      []service.Router
	cfg          *Cfg

	lc           lifecycle.GetLifecycle
	cmd          *cli.Command
	httpSrv      fiber_builder2.Builder
	providerList []interface{}
}

func (t *serviceImpl) Start() error { return t.start() }
func (t *serviceImpl) Stop() error  { return t.stop() }

func (t *serviceImpl) DixInject(cfg *Cfg, _ *registry.Loader, app *fiber2.App) {
	t.cfg = cfg

}

func (t *serviceImpl) init() (gErr error) {
	defer xerror.RecoverErr(&gErr)

	dix.Inject(t)

	t.Log = t.Log.Named(runmode.Project)

	var middlewares []service.Middleware
	for _, m := range t.Middlewares {
		middlewares = append(middlewares, m)
	}

	// 网关初始化
	xerror.Panic(t.httpSrv.Build(t.cfg.Api))
	t.httpSrv.Get().Use(t.handlerHttpMiddle(middlewares))

	for i := range t.Routers {
		t.Routers[i](t.httpSrv.Get())
	}

	if t.cfg.PrintRoute {
		for _, stacks := range t.httpSrv.Get().Stack() {
			for _, s := range stacks {
				t.Log.Info("service route",
					zap.String("name", s.Name),
					zap.String("path", s.Path),
					zap.String("method", s.Method),
				)
			}
		}
	}

	return nil
}

func (t *serviceImpl) SubCmd(cmd *cli.Command) {
	t.cmd.Subcommands = append(t.cmd.Subcommands, cmd)
}

func (t *serviceImpl) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
}

func (t *serviceImpl) Provider(provider interface{}) {
	t.providerList = append(t.providerList, provider)
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
		for _, run := range append(t.lc.GetBeforeStarts(), t.GetLifecycle.GetBeforeStarts()...) {
			t.Log.Sugar().Infof("before-start running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})

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
		for _, run := range append(t.lc.GetAfterStarts(), t.GetLifecycle.GetAfterStarts()...) {
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
		for _, run := range append(t.lc.GetBeforeStops(), t.GetLifecycle.GetBeforeStops()...) {
			t.Log.Sugar().Infof("before-stop running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})

	logutil.LogOrErr(t.Log, "[grpc-gw] Shutdown", func() error {
		xerror.Panic(t.httpSrv.Get().Shutdown())
		return nil
	})

	logutil.OkOrErr(t.Log, "service after-stop", func() error {
		for _, run := range append(t.lc.GetAfterStops(), t.GetLifecycle.GetAfterStops()...) {
			t.Log.Sugar().Infof("after-stop running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})

	return
}
