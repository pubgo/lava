package https

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/logx"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/syncx"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/projectinfo"
	"github.com/pubgo/lava/core/requestid"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/debug"
	"github.com/pubgo/lava/logging/logmiddleware"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/x/stack"
	"go.uber.org/zap"
)

func New() service.Service { return newService() }

func newService() *serviceImpl {
	return &serviceImpl{}
}

var _ service.Service = (*serviceImpl)(nil)

type serviceImpl struct {
	lc         lifecycle.GetLifecycle
	httpServer *fiber.App
	log        *zap.Logger
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
	log *zap.Logger,
	cfg *Cfg) {

	log = log.Named("http-server")

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

	middlewares = append([]service.Middleware{
		logmiddleware.Middleware(log),
		requestid.Middleware(),
		projectinfo.Middleware()}, middlewares...)

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
			for _, s := range stacks {
				logx.Info(
					"service route",
					"name", s.Name,
					"path", s.Path,
					"method", s.Method,
				)
			}
		}
	}
}

func (s *serviceImpl) start() {
	logutil.OkOrFailed(s.log, "service before-start", func() result.Error {
		defer recovery.Exit()
		for _, run := range s.lc.GetBeforeStarts() {
			s.log.Sugar().Infof("running %s", stack.Func(run.Handler))
			run.Handler()
		}
		return result.NilErr()
	})

	httpLn := assert.Must1(net.Listen("tcp", fmt.Sprintf(":%d", runmode.HttpPort)))

	logutil.OkOrFailed(s.log, "service handler init", func() result.Error {
		defer recovery.Exit()
		for _, ii := range s.initList {
			s.log.Sugar().Infof("handler %s", stack.Func(ii))
			ii()
		}
		return result.NilErr()
	})

	logutil.OkOrFailed(s.log, "service start", func() result.Error {
		syncx.GoDelay(func() result.Error {
			s.log.Info("[http-server] Server Starting")
			logutil.LogOrErr(s.log, "[http-server] Server Stop", func() result.Error {
				defer recovery.Exit()
				if err := s.httpServer.Listener(httpLn); err != nil &&
					!errors.Is(err, http.ErrServerClosed) &&
					!errors.Is(err, net.ErrClosed) {
					return result.WithErr(err)
				}
				return result.NilErr()
			})
			return result.NilErr()
		})
		return result.NilErr()
	})

	logutil.OkOrFailed(s.log, "service after-start", func() result.Error {
		defer recovery.Exit()
		for _, run := range s.lc.GetAfterStarts() {
			s.log.Sugar().Infof("running %s", stack.Func(run.Handler))
			run.Handler()
		}
		return result.NilErr()
	})
}

func (s *serviceImpl) stop() {
	logutil.OkOrFailed(s.log, "service before-stop", func() result.Error {
		for _, run := range s.lc.GetBeforeStops() {
			s.log.Sugar().Infof("running %s", stack.Func(run.Handler))
			run.Handler()
		}
		return result.NilErr()
	})

	logutil.LogOrErr(s.log, "[http-server] Shutdown", func() result.Error {
		return result.WithErr(s.httpServer.Shutdown())
	})

	logutil.OkOrFailed(s.log, "service after-stop", func() result.Error {
		for _, run := range s.lc.GetAfterStops() {
			s.log.Sugar().Infof("running %s", stack.Func(run.Handler))
			run.Handler()
		}
		return result.NilErr()
	})
}
