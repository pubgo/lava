package https

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/async"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/vars"
	"github.com/samber/lo"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/core/running"
	"github.com/pubgo/lava/v2/core/supervisor"
	"github.com/pubgo/lava/v2/internal/logutil"
	"github.com/pubgo/lava/v2/pkg/middleware/accesslog"
	"github.com/pubgo/lava/v2/pkg/middleware/metric"
	mwrecovery "github.com/pubgo/lava/v2/pkg/middleware/recovery"
	"github.com/pubgo/lava/v2/pkg/middleware/serviceinfo"
	"github.com/pubgo/lava/v2/pkg/lava"
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

func (s *serviceImpl) String() string { return "http-server" }
func (s *serviceImpl) Serve(ctx context.Context) error {
	defer s.stop(ctx)
	s.start(ctx)
	<-ctx.Done()
	return nil
}

func (s *serviceImpl) init(params Params) {
	cfg := lo.ToPtr(httputil.DefaultCfg(params.Cfg))

	s.log = params.Log.WithName(s.String())
	s.httpServer = fiber.New(cfg.Http.Build().Unwrap())
	s.httpServer.Use(httputil.Cors())
	s.httpServer.Use("/debug", debug.App())

	defaultMiddlewares := make([]lava.Middleware, 0, 4+len(params.Middlewares))
	defaultMiddlewares = append(defaultMiddlewares,
		serviceinfo.New(),
		metric.New(params.M),
		accesslog.New(s.log),
		mwrecovery.New(),
	)
	middlewares := append(defaultMiddlewares, params.Middlewares...)

	for _, h := range params.Handlers {
		h.Router(s.httpServer.Group(h.Prefix(), handlerHttpMiddle(append(middlewares, h.Middlewares()...))))
	}

	vars.Register(vars.UniqueName(running.Project(), "http_server_info"), func() any {
		return map[string]any{
			"config": cfg,
			"router": s.httpServer.Stack(),
		}
	})
}

func (s *serviceImpl) start(ctx context.Context) {
	defer recovery.Exit()

	addr := fmt.Sprintf(":%d", running.HttpPort.Value())
	async.GoDelay(func() error {
		defer recovery.Exit()

		s.log.Info().Msg("http server starting")
		err := s.httpServer.Listen(addr)
		if netutil.IsErrServerClosed(err) {
			return nil
		}

		return err
	})
}

func (s *serviceImpl) stop(ctx context.Context) {
	defer recovery.DebugPrint()
	logutil.LogOrErr(s.log, "http server shutdown", func() error {
		err := s.httpServer.ShutdownWithContext(ctx)
		if netutil.IsErrServerClosed(err) {
			return nil
		}

		return err
	})
}
