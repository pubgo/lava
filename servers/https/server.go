package https

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/errors/errutil"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/opendoc/opendoc"
	"google.golang.org/grpc/codes"

	"github.com/pubgo/lava/core/debug"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/internal/logutil"
	"github.com/pubgo/lava/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/internal/middlewares/middleware_service_info"
	"github.com/pubgo/lava/lava"
)

func New() lava.Server { return newService() }

func newService() *serviceImpl {
	return &serviceImpl{}
}

var _ lava.Server = (*serviceImpl)(nil)

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

func (s *serviceImpl) DixInject(
	handlers []lava.HttpRouter,
	middlewares []lava.Middleware,
	getLifecycle lifecycle.Getter,
	lifecycle lifecycle.Lifecycle,
	m metrics.Metric,
	log log.Logger,
	cfg *Config,
	docs []*opendoc.Swagger,
) {
	if cfg.BaseUrl == "" {
		cfg.BaseUrl = "/" + version.Project()
	}

	log = log.WithName("http-server")

	s.lc = getLifecycle
	s.log = log

	s.httpServer = fiber.New(fiber.Config{
		EnableIPValidation: true,
		ETag:               true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			if err == nil {
				return nil
			}

			errPb := errutil.ParseError(err)
			if errPb == nil || errPb.Code.Code == 0 {
				return nil
			}

			errPb.Trace.Operation = ctx.Route().Path
			code := errutil.GrpcCodeToHTTP(codes.Code(errPb.Code.Code))
			ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			return ctx.Status(code).JSON(errPb)
		},
	})

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return true
		},
		AllowOrigins: "*",
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
			fiber.MethodHead,
			fiber.MethodOptions,
		}, ","),
		AllowHeaders:     "",
		AllowCredentials: true,
		ExposeHeaders:    "",
		MaxAge:           0,
	}))

	defaultMiddlewares := []lava.Middleware{
		middleware_service_info.New(),
		middleware_metric.New(m),
		middleware_accesslog.New(log),
		middleware_recovery.New(),
	}
	app.Use(handlerHttpMiddle(append(defaultMiddlewares, middlewares...)))

	for _, h := range handlers {
		g := app.Group("", handlerHttpMiddle(h.Middlewares()))

		//for _, an := range h.Annotation() {
		//	switch a := an.(type) {
		//	case *annotation.Openapi:
		//		if a.ServiceName != "" {
		//			srv.SetName(a.ServiceName)
		//		}
		//	}
		//}

		h.Router(g)

		if m, ok := h.(lava.Close); ok {
			lifecycle.BeforeStop(m.Close)
		}
	}

	s.httpServer.Mount("/debug", debug.App())
	s.httpServer.Mount(cfg.BaseUrl, app)

	// 网关初始化
	if cfg.EnablePrintRouter {
		for _, stacks := range s.httpServer.Stack() {
			for _, route := range stacks {
				s.log.Info().
					Str("name", route.Name).
					Str("path", route.Path).
					Str("method", route.Method).
					Msg("service route")
			}
		}
	}
}

func (s *serviceImpl) start(ctx context.Context) {
	defer recovery.Exit()
	logutil.OkOrFailed(s.log, "service before-start", func() error {
		defer recovery.Exit()
		for _, run := range s.lc.GetBeforeStarts() {
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Exec))
			assert.Exit(run.Exec(ctx))
		}
		return nil
	})

	httpLn := assert.Must1(net.Listen("tcp", fmt.Sprintf(":%d", running.HttpPort)))

	logutil.OkOrFailed(s.log, "service start", func() error {
		async.GoDelay(func() error {
			s.log.Info().Msg("[http-server] Server Starting")
			logutil.LogOrErr(s.log, "[http-server] Server Stop", func() error {
				defer recovery.Exit()
				if err := s.httpServer.Listener(httpLn); err != nil &&
					!errors.Is(err, http.ErrServerClosed) &&
					!errors.Is(err, net.ErrClosed) {
					return err
				}
				return nil
			})
			return nil
		})
		return nil
	})

	logutil.OkOrFailed(s.log, "service after-start", func() error {
		defer recovery.Exit()
		for _, run := range s.lc.GetAfterStarts() {
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Exec))
			assert.Exit(run.Exec(ctx))
		}
		return nil
	})
}

func (s *serviceImpl) stop(ctx context.Context) {
	defer recovery.DebugPrint()
	logutil.OkOrFailed(s.log, "service before-stop", func() error {
		for _, run := range s.lc.GetBeforeStops() {
			logutil.LogOrErr(s.log, fmt.Sprintf("running %s", stack.CallerWithFunc(run.Exec)), func() error {
				return run.Exec(ctx)
			})
		}
		return nil
	})

	logutil.LogOrErr(s.log, "[http-server] Shutdown", func() error {
		return s.httpServer.ShutdownWithTimeout(time.Second * 5)
	})

	logutil.OkOrFailed(s.log, "service after-stop", func() error {
		for _, run := range s.lc.GetAfterStops() {
			logutil.LogOrErr(s.log, fmt.Sprintf("running %s", stack.CallerWithFunc(run.Exec)), func() error {
				return run.Exec(ctx)
			})
		}
		return nil
	})
}
