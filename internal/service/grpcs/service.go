package grpcs

import (
	"errors"
	"net"
	"net/http"

	"github.com/fullstorydev/grpchan"
	"github.com/fullstorydev/grpchan/httpgrpc"
	"github.com/gofiber/adaptor/v2"
	fiber2 "github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/runmode"
	cmux2 "github.com/pubgo/lava/internal/cmux"
	fiber_builder2 "github.com/pubgo/lava/internal/pkg/fiber_builder"
	grpc_builder2 "github.com/pubgo/lava/internal/pkg/grpc_builder"
	netutil2 "github.com/pubgo/lava/internal/pkg/netutil"
	"github.com/pubgo/lava/internal/pkg/syncx"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/version"
)

func New(name string, desc ...string) service.Service {
	return newService(name, desc...)
}

func newService(name string, desc ...string) *serviceImpl {
	return &serviceImpl{
		grpcSrv:  grpc_builder2.New(),
		httpSrv:  fiber_builder2.New(),
		handlers: grpchan.HandlerMap{},
	}
}

var _ service.Service = (*serviceImpl)(nil)

type serviceImpl struct {
	getLifecycle lifecycle.GetLifecycle
	lc           lifecycle.Lifecycle
	net          *cmux2.Mux
	app          *service.WebApp
	cfg          *Cfg
	log          *zap.Logger
	grpcSrv      grpc_builder2.Builder
	httpSrv      fiber_builder2.Builder
	providerList []interface{}
	handlers     grpchan.HandlerMap

	unaryInt   grpc.UnaryServerInterceptor
	streamInt  grpc.StreamServerInterceptor
	httpMiddle func(_ *fiber2.Ctx) error
}

func (t *serviceImpl) Start() error { return t.start() }
func (t *serviceImpl) Stop() error  { return t.stop() }

func (t *serviceImpl) init() (gErr error) {
	defer xerror.RecoverErr(&gErr)

	type injectParam struct {
		Middlewares  []service.Middleware
		GetLifecycle lifecycle.GetLifecycle
		Lifecycle    lifecycle.Lifecycle
		Log          *zap.Logger
		Net          *cmux2.Mux
		App          *service.WebApp
		Cfg          *Cfg
	}

	dix.Inject(func(p injectParam) {
		t.getLifecycle = p.GetLifecycle
		t.lc = p.Lifecycle
		t.log = p.Log.Named(runmode.Project)
		t.net = p.Net
		t.app = p.App
		t.cfg = p.Cfg

		var middlewares []service.Middleware
		for _, m := range p.Middlewares {
			middlewares = append(middlewares, m)
		}

		t.unaryInt = t.handlerUnaryMiddle(middlewares)
		t.streamInt = t.handlerStreamMiddle(middlewares)
		t.httpMiddle = t.handlerHttpMiddle(middlewares)
	})

	t.handlers.ForEach(func(_ *grpc.ServiceDesc, svc interface{}) {
		dix.Inject(svc)
	})

	// 网关初始化
	xerror.Panic(t.httpSrv.Build(t.cfg.Api))
	t.httpSrv.Get().Use(t.httpMiddle)
	t.httpSrv.Get().Mount("/", t.app.App)

	httpgrpc.HandleServices(func(pattern string, handler func(http.ResponseWriter, *http.Request)) {
		t.httpSrv.Get().Options(pattern, func(ctx *fiber2.Ctx) error {
			ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
			ctx.Response().Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			ctx.Response().Header.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, X-Extra-Header, Content-Type, Accept, Authorization")
			ctx.Response().Header.Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			return ctx.SendStatus(http.StatusOK)
		})

		t.httpSrv.Get().Post(pattern, func(ctx *fiber2.Ctx) error {
			ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
			ctx.Response().Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			ctx.Response().Header.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, X-Extra-Header, Content-Type, Accept, Authorization")
			ctx.Response().Header.Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			return adaptor.HTTPHandlerFunc(handler)(ctx)
		})
	}, "/"+runmode.Project, t.handlers, t.unaryInt, t.streamInt)

	// 注册系统middleware
	t.grpcSrv.UnaryInterceptor(t.unaryInt)
	t.grpcSrv.StreamInterceptor(t.streamInt)

	// grpc serve初始化
	xerror.Panic(t.grpcSrv.Build(t.cfg.Grpc))

	// 初始化 handlers
	t.handlers.ForEach(func(desc *grpc.ServiceDesc, svr interface{}) {
		t.grpcSrv.Get().RegisterService(desc, svr)

		if h, ok := svr.(service.Close); ok {
			t.lc.AfterStops(h.Close)
		}

		if h, ok := svr.(service.Init); ok {
			h.Init()
		}

		if h, ok := svr.(service.WebHandler); ok {
			h.Router(t.httpSrv.Get())
		}
	})

	if t.cfg.PrintRoute {
		for _, stacks := range t.httpSrv.Get().Stack() {
			for _, s := range stacks {
				t.log.Info("service route",
					zap.String("name", s.Name),
					zap.String("path", s.Path),
					zap.String("method", s.Method),
				)
			}
		}
	}

	return nil
}

func (t *serviceImpl) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	xerror.Assert(desc == nil, "[desc] is nil")
	xerror.Assert(impl == nil, "[impl] is nil")
	t.handlers.RegisterService(desc, impl)
}

func (t *serviceImpl) Provider(provider interface{}) {
	dix.Provider(provider)
}

func (t *serviceImpl) Options() service.Options {
	return service.Options{
		Name:      runmode.Project,
		Id:        runmode.InstanceID,
		Version:   version.Version(),
		Port:      netutil2.MustGetPort(t.net.Addr),
		Addr:      t.net.Addr,
		Advertise: "",
	}
}

func (t *serviceImpl) start() (gErr error) {
	defer xerror.RecoverErr(&gErr)

	xerror.Panic(t.init())

	logutil.OkOrPanic(t.log, "service before-start", func() error {
		for _, run := range t.getLifecycle.GetBeforeStarts() {
			t.log.Sugar().Infof("before-start running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})

	var grpcLn = t.net.HTTP2()
	var gwLn = t.net.HTTP1()

	logutil.OkOrPanic(t.log, "service start", func() error {
		t.log.Sugar().Infof("Server Listening on http://%s:%d", netutil2.GetLocalIP(), netutil2.MustGetPort(t.net.Addr))

		// 启动grpc网关
		syncx.GoDelay(func() {
			t.log.Info("[grpc-gw] Server Starting")
			logutil.LogOrErr(t.log, "[grpc-gw] Server Stop", func() error {
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
			t.log.Info("[grpc] Server Starting")
			logutil.LogOrErr(t.log, "[grpc] Server Stop", func() error {
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
		for _, run := range t.getLifecycle.GetAfterStarts() {
			t.log.Sugar().Infof("after-start running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})
	return nil
}

func (t *serviceImpl) stop() (err error) {
	defer xerror.RecoverErr(&err)

	logutil.OkOrErr(t.log, "service before-stop", func() error {
		for _, run := range t.getLifecycle.GetBeforeStops() {
			t.log.Sugar().Infof("before-stop running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})

	logutil.LogOrErr(t.log, "[grpc-gw] Shutdown", func() error {
		xerror.Panic(t.httpSrv.Get().Shutdown())
		return nil
	})

	logutil.LogOrErr(t.log, "[grpc] GracefulStop", func() error {
		t.grpcSrv.Get().GracefulStop()
		return t.net.Close()
	})

	logutil.OkOrErr(t.log, "service after-stop", func() error {
		for _, run := range t.getLifecycle.GetAfterStops() {
			t.log.Sugar().Infof("after-stop running %s", stack.Func(run))
			xerror.PanicF(xerror.Try(run), stack.Func(run))
		}
		return nil
	})

	return
}
