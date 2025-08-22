package https

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/errors/errutil"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/proto/errorpb"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/opendoc/opendoc"
	"google.golang.org/grpc/codes"
	"net"
	"net/http"
	"strings"

	"github.com/pubgo/lava/core/debug"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/core/supervisor"
	"github.com/pubgo/lava/internal/logutil"
	"github.com/pubgo/lava/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/internal/middlewares/middleware_serviceinfo"
	"github.com/pubgo/lava/lava"
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

	return supervisor.NewService("http-server", s.Serve)
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
	if cfg.BaseUrl == "" {
		cfg.BaseUrl = "/" + version.Project()
	}

	log = log.WithName("http-server")

	s.log = log

	s.httpServer = fiber.New(fiber.Config{
		EnableIPValidation: true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			if err == nil {
				return nil
			}

			var errPb *errorpb.Error
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) && fiberErr != nil {
				errPb = &errorpb.Error{
					Code: &errorpb.ErrCode{
						Name:       "lava.error",
						StatusCode: errorpb.Code(errutil.Http2GrpcCode(int32(fiberErr.Code))),
						Code:       int32(fiberErr.Code),
						Message:    fiberErr.Message,
					},
					Trace: &errorpb.ErrTrace{},
				}
			} else {
				errPb = errutil.ParseError(err)
			}

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
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
			fiber.MethodHead,
			fiber.MethodOptions,
		}, ","),
		//AllowHeaders:     "",
		AllowCredentials: true,
		//ExposeHeaders:    "",
		MaxAge: 0,
	}))

	defaultMiddlewares := []lava.Middleware{
		middleware_serviceinfo.New(),
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
}

func (s *serviceImpl) stop(ctx context.Context) {
	defer recovery.DebugPrint()
	logutil.LogOrErr(s.log, "[http-server] Shutdown", func() error {
		return s.httpServer.ShutdownWithContext(ctx)
	})
}
