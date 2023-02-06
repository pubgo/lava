package https

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/debug"
	"github.com/pubgo/funk/lifecycle"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/log/logutil"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/stack"

	"github.com/pubgo/lava/core/projectinfo"
	"github.com/pubgo/lava/core/requestid"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/logging/logmiddleware"
	"github.com/pubgo/lava/service"
)

func New() service.Service { return newService() }

func newService() *serviceImpl {
	return &serviceImpl{}
}

var _ service.Service = (*serviceImpl)(nil)

type serviceImpl struct {
	lc         lifecycle.GetLifecycle
	httpServer *fiber.App
	log        log.Logger
	initList   []func()
}

func (s *serviceImpl) Run() {
	defer s.stop()
	s.start()
	signal.Wait()
}

func (s *serviceImpl) Start() { s.start() }
func (s *serviceImpl) Stop()  { s.stop() }

func (s *serviceImpl) DixInject(
	handlers []service.HttpRouter,
	middlewares []service.Middleware,
	getLifecycle lifecycle.GetLifecycle,
	lifecycle lifecycle.Lifecycle,
	log log.Logger,
	cfg *Cfg) {

	log = log.WithName("http-server")

	s.lc = getLifecycle
	s.log = log

	s.httpServer = fiber.New(fiber.Config{
		EnableIPValidation: true,
	})

	s.httpServer.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowCredentials: true,
	}))
	s.httpServer.Mount("/debug", debug.App())

	app := fiber.New()

	var defaultMiddlewares = []service.Middleware{logmiddleware.Middleware(log), requestid.Middleware(), projectinfo.Middleware()}
	middlewares = append(defaultMiddlewares, middlewares...)

	for _, h := range handlers {
		s.initList = append(s.initList, h.Init)
		middlewares = append(middlewares, h.Middlewares()...)

		h.Router(app)

		if m, ok := h.(service.Close); ok {
			lifecycle.BeforeStop(m.Close)
		}
	}

	app.Use(handlerHttpMiddle(middlewares))

	s.httpServer.Mount("/", app)

	// 网关初始化
	if cfg.PrintRoute {
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

func (s *serviceImpl) start() {
	logutil.OkOrFailed(s.log, "service before-start", func() error {
		defer recovery.Exit()
		for _, run := range s.lc.GetBeforeStarts() {
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Handler))
			run.Handler()
		}
		return nil
	})

	httpLn := assert.Must1(net.Listen("tcp", fmt.Sprintf(":%d", runmode.HttpPort)))

	logutil.OkOrFailed(s.log, "service handler init", func() error {
		defer recovery.Exit()
		for _, ii := range s.initList {
			s.log.Info().Msgf("handler %s", stack.CallerWithFunc(ii))
			ii()
		}
		return nil
	})

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
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Handler))
			run.Handler()
		}
		return nil
	})
}

func (s *serviceImpl) stop() {
	logutil.OkOrFailed(s.log, "service before-stop", func() error {
		for _, run := range s.lc.GetBeforeStops() {
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Handler))
			run.Handler()
		}
		return nil
	})

	logutil.LogOrErr(s.log, "[http-server] Shutdown", func() error {
		return s.httpServer.Shutdown()
	})

	logutil.OkOrFailed(s.log, "service after-stop", func() error {
		for _, run := range s.lc.GetAfterStops() {
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Handler))
			run.Handler()
		}
		return nil
	})
}
