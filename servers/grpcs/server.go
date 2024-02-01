package grpcs

import (
	"errors"
	"fmt"
	"github.com/pubgo/lava/core/annotation"
	"net"
	"net/http"
	"net/url"
	"strings"

	_ "github.com/fullstorydev/grpchan/httpgrpc"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/errors/errutil"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/funk/vars"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/pkg/gateway"
	"github.com/pubgo/lava/pkg/grpcutil"
	"github.com/pubgo/lava/pkg/httputil"
	"github.com/pubgo/opendoc/opendoc"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/pubgo/lava/core/debug"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/internal/consts"
	"github.com/pubgo/lava/internal/logutil"
	"github.com/pubgo/lava/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/internal/middlewares/middleware_service_info"
	"github.com/pubgo/lava/lava"
)

func New() lava.Service { return newService() }

func newService() *serviceImpl {
	return &serviceImpl{}
}

var _ lava.Service = (*serviceImpl)(nil)

type serviceImpl struct {
	lc         lifecycle.Getter
	httpServer *fiber.App
	grpcServer *grpc.Server
	log        log.Logger
	initList   []func()
	conf       *Config
}

func (s *serviceImpl) Run() {
	defer s.stop()
	s.start()
	signal.Wait()
}

func (s *serviceImpl) Start() { s.start() }
func (s *serviceImpl) Stop()  { s.stop() }

func (s *serviceImpl) DixInject(
	handlers []lava.GrpcRouter,
	httpRouters []lava.HttpRouter,
	dixMiddlewares []lava.Middleware,
	getLifecycle lifecycle.Getter,
	lifecycle lifecycle.Lifecycle,
	metric metrics.Metric,
	log log.Logger,
	conf *Config,
	docs []*opendoc.Swagger,
	empty []*lava.EmptyRouter,
) {
	_ = empty
	s.conf = conf
	if conf.HttpPort == nil {
		conf.HttpPort = generic.Ptr(running.HttpPort)
	}

	if conf.GrpcPort == nil {
		conf.GrpcPort = generic.Ptr(running.GrpcPort)
	}

	if conf.BaseUrl == "" {
		conf.BaseUrl = "/" + version.Project()
	}

	fiber.SetParserDecoder(fiber.ParserConfig{
		IgnoreUnknownKeys: true,
		ZeroEmpty:         true,
		ParserType:        parserTypes,
	})

	s.lc = getLifecycle

	conf = config.MergeR(defaultCfg(), conf).Unwrap()
	conf.BaseUrl = "/" + strings.Trim(conf.BaseUrl, "/")

	var doc = opendoc.New(func(swag *opendoc.Swagger) {
		swag.Config.Title = "service title "
		swag.Description = "this is description"
		swag.License = &opendoc.License{
			Name: "Apache License 2.0",
			URL:  "https://github.com/pubgo/opendoc/blob/master/LICENSE",
		}

		swag.Contact = &opendoc.Contact{
			Name:  "barry",
			URL:   "https://github.com/pubgo/opendoc",
			Email: "kooksee@163.com",
		}

		swag.TermsOfService = "https://github.com/pubgo"
	})
	if len(docs) > 0 {
		doc = docs[0]
	}
	doc.SetRootPath(conf.BaseUrl)

	middlewares := lava.Middlewares{
		middleware_service_info.New(),
		middleware_metric.New(metric),
		middleware_accesslog.New(log),
		middleware_recovery.New(),
	}
	middlewares = append(middlewares, dixMiddlewares...)

	log = log.WithName("grpc-server")
	s.log = log

	httpServer := fiber.New(fiber.Config{
		EnableIPValidation: true,
		EnablePrintRoutes:  conf.EnablePrintRoutes,
		AppName:            version.Project(),
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			if err == nil {
				return nil
			}

			code := fiber.StatusBadRequest
			errPb := errutil.ParseError(err)
			if errPb == nil || errPb.Code.Code == 0 {
				return nil
			}

			errPb.Trace.Operation = ctx.Route().Path
			code = errutil.GrpcCodeToHTTP(codes.Code(errPb.Code.Code))
			ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			return ctx.Status(code).JSON(errPb.Code)
		},
	})

	if conf.EnableCors {
		httpServer.Use(cors.New(cors.Config{
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
	}

	httpServer.Mount("/debug", debug.App())

	app := fiber.New()
	app.Use(handlerHttpMiddle(middlewares))
	for _, h := range httpRouters {
		srv := doc.WithService()
		for _, an := range h.Annotation() {
			switch a := an.(type) {
			case *annotation.Openapi:
				if a.ServiceName != "" {
					srv.SetName(a.ServiceName)
				}
			}
		}

		h.Router(&lava.Router{R: app.Group("", handlerHttpMiddle(h.Middlewares())), Doc: srv})

		if m, ok := h.(lava.Init); ok {
			s.initList = append(s.initList, m.Init)
		}

		if m, ok := h.(lava.Initializer); ok {
			s.initList = append(s.initList, m.Initialize)
		}
	}

	httpServer.Mount(conf.BaseUrl, app)

	var mux = gateway.NewMux()
	srvMidMap := make(map[string][]lava.Middleware)
	for _, h := range handlers {
		desc := h.ServiceDesc()
		assert.If(desc == nil, "desc is nil")

		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], middlewares...)
		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], h.Middlewares()...)

		if m, ok := h.(lava.Initializer); ok {
			s.initList = append(s.initList, m.Initialize)
		}

		if m, ok := h.(lava.Init); ok {
			s.initList = append(s.initList, m.Init)
		}

		mux.RegisterService(desc, h)
	}

	mux.WithServerUnaryInterceptor(handlerUnaryMiddle(srvMidMap))
	mux.WithServerStreamInterceptor(handlerStreamMiddle(srvMidMap))

	// grpc server初始化
	grpcServer := conf.GrpcConfig.Build(
		grpc.ChainUnaryInterceptor(handlerUnaryMiddle(srvMidMap)),
		grpc.ChainStreamInterceptor(handlerStreamMiddle(srvMidMap))).Unwrap()

	for _, h := range handlers {
		grpcServer.RegisterService(h.ServiceDesc(), h)
	}

	wrappedGrpc := grpcweb.WrapServer(grpcServer,
		grpcweb.WithWebsockets(true),
		grpcweb.WithAllowNonRootResource(true),
		grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool { return true }),
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithOriginFunc(func(origin string) bool { return true }))

	apiPrefix1 := assert.Must1(url.JoinPath(conf.BaseUrl, "api"))
	httpServer.Group(apiPrefix1, httputil.StripPrefix(apiPrefix1, func(ctx *fiber.Ctx) error {
		mux.GetApp().Handler()(ctx.Context())
		return nil
	}))

	grpcWebApiPrefix := assert.Must1(url.JoinPath(conf.BaseUrl, "grpc"))
	s.log.Info().Str("path", grpcWebApiPrefix).Msg("service grpc web base path")
	httpServer.Group(grpcWebApiPrefix+"/*", adaptor.HTTPHandler(h2c.NewHandler(http.StripPrefix(grpcWebApiPrefix,
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if wrappedGrpc.IsAcceptableGrpcCorsRequest(request) {
				writer.WriteHeader(http.StatusNoContent)
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

			if grpcutil.IsGRPCRequest(request) {
				grpcServer.ServeHTTP(writer, request)
			}
		})), new(http2.Server))))

	s.httpServer = httpServer
	s.grpcServer = grpcServer

	vars.RegisterValue(fmt.Sprintf("%s-grpc-server-config", version.Project()), &conf)
}

func (s *serviceImpl) start() {
	defer recovery.Exit()

	logutil.OkOrFailed(s.log, "running before service starts", func() error {
		defer recovery.Exit()
		for _, run := range s.lc.GetBeforeStarts() {
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Handler))
			run.Handler()
		}
		return nil
	})

	logutil.OkOrFailed(s.log, "init handler before service starts", func() error {
		defer recovery.Exit()
		for _, ii := range s.initList {
			s.log.Info().Msgf("init handler %s", stack.CallerWithFunc(ii))
			ii()
		}
		return nil
	})

	s.log.Info().
		Int("grpc-port", *s.conf.GrpcPort).
		Int("http-port", *s.conf.HttpPort).
		Msg("create network listener")
	grpcLn := assert.Must1(net.Listen("tcp", fmt.Sprintf(":%d", *s.conf.GrpcPort)))
	httpLn := assert.Must1(net.Listen("tcp", fmt.Sprintf(":%d", *s.conf.HttpPort)))

	logutil.OkOrFailed(s.log, "service starts", func() error {
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
			s.log.Info().Msg("[http] Server Starting")
			logutil.LogOrErr(s.log, "[http] Server Stop", func() error {
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

	logutil.OkOrFailed(s.log, "running after service starts", func() error {
		for _, run := range s.lc.GetAfterStarts() {
			logutil.LogOrErr(
				s.log,
				fmt.Sprintf("running %s", stack.CallerWithFunc(run.Handler)),
				func() error { run.Handler(); return nil },
			)
		}
		return nil
	})
}

func (s *serviceImpl) stop() {
	defer recovery.Exit()

	logutil.OkOrFailed(s.log, "running before service stops", func() error {
		for _, run := range s.lc.GetBeforeStops() {
			logutil.LogOrErr(
				s.log,
				fmt.Sprintf("running %s", stack.CallerWithFunc(run.Handler)),
				func() error { run.Handler(); return nil },
			)
		}
		return nil
	})

	logutil.LogOrErr(s.log, "[grpc] Server GracefulStop", func() error {
		s.grpcServer.GracefulStop()
		return nil
	})

	logutil.LogOrErr(s.log, "[http] Server Shutdown", func() error {
		return s.httpServer.ShutdownWithTimeout(consts.DefaultTimeout)
	})

	logutil.OkOrFailed(s.log, "running after service stops", func() error {
		for _, run := range s.lc.GetAfterStops() {
			logutil.LogOrErr(
				s.log,
				fmt.Sprintf("running %s", stack.CallerWithFunc(run.Handler)),
				func() error { run.Handler(); return nil },
			)
		}
		return nil
	})
}
