package grpcs

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/debug"
	"github.com/pubgo/funk/lifecycle"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/log/logutil"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/runmode"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/funk/version"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/core/projectinfo"
	"github.com/pubgo/lava/core/requestid"
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
	grpcServer *grpc.Server
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
	handlers []service.GrpcHandler,
	middlewares []service.Middleware,
	getLifecycle lifecycle.GetLifecycle,
	lifecycle lifecycle.Lifecycle,
	log log.Logger,
	cfg *Config,
) {

	pathPrefix := "/" + strings.Trim(cfg.PathPrefix, "/")
	middlewares = append([]service.Middleware{
		logmiddleware.Middleware(log),
		requestid.Middleware(),
		projectinfo.Middleware(),
	}, middlewares...)

	log = log.WithName("grpc-server")

	httpServer := fiber.New(fiber.Config{
		EnableIPValidation: true,
		EnablePrintRoutes:  cfg.PrintRoute,
		AppName:            version.Project(),
	})
	httpServer.Mount("/debug", debug.App())
	httpServer.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowCredentials: true,
	}))

	var srvMidMap = make(map[string][]service.Middleware)
	for _, h := range handlers {
		desc := h.ServiceDesc()
		assert.If(desc == nil, "desc is nil")

		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], middlewares...)
		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], h.Middlewares()...)

		s.initList = append(s.initList, h.Init)
		if m, ok := h.(service.Close); ok {
			lifecycle.BeforeStop(m.Close)
		}
	}

	// grpc server初始化
	var grpcServer = cfg.Grpc.Build(
		grpc.ChainUnaryInterceptor(handlerUnaryMiddle(srvMidMap)),
		grpc.ChainStreamInterceptor(handlerStreamMiddle(srvMidMap))).Unwrap()

	for _, h := range handlers {
		basePrefix := filepath.Join(pathPrefix, h.ServiceDesc().ServiceName, "*")
		s.httpServer.Post(basePrefix, handlerTwMiddle(srvMidMap, h.Gateway(nil)))
		grpcServer.RegisterService(h.ServiceDesc(), h)
	}

	s.lc = getLifecycle
	s.log = log
	s.httpServer = httpServer
	s.grpcServer = grpcServer
}

func (s *serviceImpl) start() {
	defer recovery.Exit()

	logutil.OkOrFailed(s.log, "service before-start", func() error {
		defer recovery.Exit()
		for _, run := range s.lc.GetBeforeStarts() {
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Handler))
			run.Handler()
		}
		return nil
	})

	grpcLn := assert.Must1(net.Listen("tcp", fmt.Sprintf(":%d", runmode.GrpcPort)))
	httpLn := assert.Must1(net.Listen("tcp", fmt.Sprintf(":%d", runmode.HttpPort)))

	logutil.OkOrFailed(s.log, "service handler init", func() error {
		defer recovery.Exit()
		for _, ii := range s.initList {
			s.log.Info().Msgf("init handler %s", stack.CallerWithFunc(ii))
			ii()
		}
		return nil
	})

	logutil.OkOrFailed(s.log, "service start", func() error {
		// 启动grpc服务
		async.GoDelay(func() error {
			s.log.Info().Msg("[grpc] Server Starting")
			logutil.LogOrErr(s.log, "[grpc] Server Stop", func() error {
				defer recovery.Exit()
				if err := s.grpcServer.Serve(grpcLn); err != nil &&
					!errors.Is(err, http.ErrServerClosed) &&
					!errors.Is(err, net.ErrClosed) {
					return err
				}
				return nil
			})
			return nil
		})

		// 启动grpc网关
		async.GoDelay(func() error {
			s.log.Info().Msg("[grpc-gw] Server Starting")
			logutil.LogOrErr(s.log, "[grpc-gw] Server Stop", func() error {
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
	defer recovery.Exit()

	logutil.OkOrFailed(s.log, "service before-stop", func() error {
		for _, run := range s.lc.GetBeforeStops() {
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Handler))
			run.Handler()
		}
		return nil
	})

	logutil.LogOrErr(s.log, "[grpc-gateway] Shutdown", func() error {
		return s.httpServer.Shutdown()
	})

	logutil.LogOrErr(s.log, "[grpc] GracefulStop", func() error {
		s.grpcServer.GracefulStop()
		return nil
	})

	logutil.OkOrFailed(s.log, "service after-stop", func() error {
		for _, run := range s.lc.GetAfterStops() {
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Handler))
			run.Handler()
		}
		return nil
	})
}
