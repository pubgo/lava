package grpcs

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/debug"
	"github.com/pubgo/funk/lifecycle"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/log/logutil"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/funk/version"
	"github.com/twitchtv/twirp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

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
	cfg *Cfg) {

	middlewares = append([]service.Middleware{
		logmiddleware.Middleware(log),
		requestid.Middleware(),
		projectinfo.Middleware(),
	}, middlewares...)

	log = log.WithName("grpc-server")

	s.lc = getLifecycle
	s.log = log

	s.httpServer = fiber.New(fiber.Config{
		EnableIPValidation: true,
		EnablePrintRoutes:  true,
		AppName:            version.Project(),
	})

	s.httpServer.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowCredentials: true,
	}))
	s.httpServer.Mount("/debug", debug.App())

	app := fiber.New()
	for _, h := range handlers {
		desc := h.ServiceDesc()
		assert.If(desc == nil, "desc is nil")

		s.initList = append(s.initList, h.Init)

		middlewares = append(middlewares, h.Middlewares()...)
		var hh = h.TwirpHandler(twirp.WithServerPathPrefix(cfg.BasePrefix))
		app.Post(cfg.BasePrefix+desc.ServiceName+"/*", adaptor.HTTPHandler(hh))

		if m, ok := h.(service.Close); ok {
			lifecycle.BeforeStop(m.Close)
		}
	}

	app.Use(handlerHttpMiddle(middlewares))

	// grpc server初始化
	var grpcServer = cfg.Grpc.Build(
		grpc.ChainUnaryInterceptor(handlerUnaryMiddle(middlewares)),
		grpc.ChainStreamInterceptor(handlerStreamMiddle(middlewares)),
	).Unwrap()
	s.grpcServer = grpcServer

	for _, h := range handlers {
		grpcServer.RegisterService(h.ServiceDesc(), h)
	}

	wrappedGrpc := grpcweb.WrapServer(grpcServer,
		grpcweb.WithWebsockets(true),
		grpcweb.WithAllowNonRootResource(true),
		grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool { return true }),
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithOriginFunc(func(origin string) bool { return true }))

	var grpcWebPrefix = "/" + version.Project() + "/grpcweb"
	var grpcPrefix = "/" + version.Project() + "/grpcweb"
	app.Post(grpcWebPrefix+"/*", adaptor.HTTPHandler(http.StripPrefix(grpcWebPrefix, wrappedGrpc)))
	app.Post(grpcPrefix+"/*", adaptor.HTTPHandler(h2c.NewHandler(http.StripPrefix(grpcPrefix, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if wrappedGrpc.IsAcceptableGrpcCorsRequest(request) {
			writer.WriteHeader(http.StatusOK)
			return
		}

		if wrappedGrpc.IsGrpcWebSocketRequest(request) {
			wrappedGrpc.HandleGrpcWebsocketRequest(writer, request)
			return
		}

		if wrappedGrpc.IsGrpcWebRequest(request) {
			wrappedGrpc.HandleGrpcWebRequest(writer, request)
			return
		}

		grpcServer.ServeHTTP(writer, request)
	})), &http2.Server{})))
	s.httpServer.Mount("/", app)

	// 网关初始化
	if cfg.PrintRoute {
		for _, stacks := range s.httpServer.Stack() {
			for _, route := range stacks {
				log.Info().
					Str("name", route.Name).
					Str("path", route.Path).
					Str("method", route.Method).
					Msg("service route")
			}
		}
	}
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
			s.log.Info().Msgf("handler %s", stack.CallerWithFunc(ii))
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
