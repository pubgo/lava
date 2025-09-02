package https

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/vars"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/opendoc/opendoc"
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
	Docs        []*opendoc.Swagger
}

func New(params Params) supervisor.Service { return newService(params) }

func newService(params Params) supervisor.Service {
	s := &serviceImpl{}
	s.init(
		params.Handlers,
		params.Middlewares,
		params.M,
		params.Log,
		params.Cfg,
		params.Docs,
	)

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

func (s *serviceImpl) init(
	handlers []lava.HttpRouter,
	middlewares []lava.Middleware,
	m metrics.Metric,
	log log.Logger,
	cfg *Config,
	docs []*opendoc.Swagger,
) {
	cfg = lo.ToPtr(httputil.DefaultCfg(cfg))

	vars.RegisterValue(s.String()+"_config_"+xid.New().String(), cfg)

	s.log = log.WithName(s.String())
	s.httpServer = fiber.New(cfg.Http.Build().Must())
	s.httpServer.Use(httputil.Cors())

	defaultMiddlewares := []lava.Middleware{
		middleware_serviceinfo.New(),
		middleware_metric.New(m),
		middleware_accesslog.New(log),
		middleware_recovery.New(),
	}
	s.httpServer.Use(handlerHttpMiddle(append(defaultMiddlewares, middlewares...)))

	for _, h := range handlers {
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

	addr := fmt.Sprintf(":%d", running.HttpPort)
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
