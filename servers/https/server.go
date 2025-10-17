package https

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/v2/async"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/running"
	"github.com/pubgo/funk/v2/vars"
	"github.com/rs/xid"
	"github.com/samber/lo"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/core/supervisor"
	"github.com/pubgo/lava/v2/internal/logutil"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_serviceinfo"
	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/httputil"
	"github.com/pubgo/lava/v2/pkg/netutil"
)

type Params struct {
	Handlers    []lava.HttpRouter
	Middlewares []lava.Middleware
	M           metrics.Metric
	Log         log.Logger
	Cfg         *Config
}

func New(params Params) supervisor.Service { return newService(params) }

func newService(params Params) supervisor.Service {
	s := &serviceImpl{}
	s.init(params)

	return supervisor.NewService(s.String(), s.Serve)
}

type serviceImpl struct {
	lc         lifecycle.Getter
	httpServer *fiber.App
	log        log.Logger
}

func (s *serviceImpl) String() string {
	return "http-server"
}

func (s *serviceImpl) Serve(ctx context.Context) error {
	defer s.stop(ctx)
	s.start(ctx)
	<-ctx.Done()
	return nil
}

func (s *serviceImpl) init(params Params) {
	cfg := lo.ToPtr(httputil.DefaultCfg(params.Cfg))

	vars.Register(s.String()+"_config_"+xid.New().String(), func() any { return cfg })

	s.log = params.Log.WithName(s.String())
	s.httpServer = fiber.New(cfg.Http.Build().Must())
	s.httpServer.Use(httputil.Cors())

	defaultMiddlewares := []lava.Middleware{
		middleware_serviceinfo.New(),
		middleware_metric.New(params.M),
		middleware_accesslog.New(s.log),
		middleware_recovery.New(),
	}
	s.httpServer.Use(handlerHttpMiddle(append(defaultMiddlewares, params.Middlewares...)))

	for _, h := range params.Handlers {
		g := s.httpServer.Group("", handlerHttpMiddle(h.Middlewares()))

		//for _, an := range h.Annotation() {
		//	switch a := an.(type) {
		//	case *annotation.Openapi:
		//		if a.ServiceName != "" {
		//			srv.SetName(a.ServiceName)
		//		}
		//	}
		//}

		h.Router(g)
	}

	s.httpServer.Mount("/debug", debug.App())

	vars.Register(fmt.Sprintf("%s-http-server-router-%s", version.Project(), xid.New()), func() interface{} {
		return s.httpServer.Stack()
	})
}

func (s *serviceImpl) start(ctx context.Context) {
	defer recovery.Exit()

	addr := fmt.Sprintf(":%d", running.HttpPort())
	async.GoDelay(func() error {
		defer recovery.Exit()

		s.log.Info().Msg("[http-server] Server Starting")
		err := s.httpServer.Listen(addr)
		if netutil.IsErrServerClosed(err) {
			return nil
		}

		return err
	})
}

func (s *serviceImpl) stop(ctx context.Context) {
	defer recovery.DebugPrint()
	logutil.LogOrErr(s.log, "[http-server] Shutdown", func() error {
		err := s.httpServer.ShutdownWithContext(ctx)
		if err == nil {
			return nil
		}

		if netutil.IsErrServerClosed(err) {
			return nil
		}

		return err
	})
}
