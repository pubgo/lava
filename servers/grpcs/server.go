package grpcs

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
	"github.com/rs/xid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/core/debug"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/internal/consts"
	"github.com/pubgo/lava/internal/logutil"
	"github.com/pubgo/lava/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/internal/middlewares/middleware_service_info"
	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/pkg/gateway"
	"github.com/pubgo/lava/pkg/httputil"
)

func New() lava.Supervisor { return newService() }

func newService() *serviceImpl {
	return &serviceImpl{
		cc: new(inprocgrpc.Channel),
	}
}

var _ lava.Supervisor = (*serviceImpl)(nil)

type serviceImpl struct {
	lc         lifecycle.Getter
	httpServer *fiber.App
	grpcServer *grpc.Server
	log        log.Logger
	cc         *inprocgrpc.Channel
	initList   []func()
	conf       *Config
}

func (s *serviceImpl) Serve(ctx context.Context) error {
	defer s.stop(ctx)
	if err := s.start(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func (s *serviceImpl) DixInject(
	grpcRouters []lava.GrpcRouter,
	httpRouters []lava.HttpRouter,
	grpcProxy []lava.GrpcProxy,
	dixMiddlewares []lava.Middleware,
	getLifecycle lifecycle.Getter,
	lifecycle lifecycle.Lifecycle,
	metric metrics.Metric,
	log log.Logger,
	conf *Config,
	gw []*gateway.Mux,
) {
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

	s.lc = getLifecycle

	conf = config.MergeR(defaultCfg(), conf).Unwrap()
	conf.BaseUrl = "/" + strings.Trim(conf.BaseUrl, "/")

	globalMiddlewares := lava.Middlewares{
		middleware_service_info.New(),
		middleware_metric.New(metric),
		middleware_accesslog.New(log),
		middleware_recovery.New(),
	}
	globalMiddlewares = append(globalMiddlewares, dixMiddlewares...)

	log = log.WithName("grpc-server")
	s.log = log

	httpServer := fiber.New(fiber.Config{
		EnableIPValidation: true,
		EnablePrintRoutes:  conf.EnablePrintRoutes,
		AppName:            version.Project(),
		BodyLimit:          500 * 1024 * 1024,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			if err == nil {
				return nil
			}

			errPb := errutil.ParseError(err)
			if errPb == nil || errPb.Code.Code == 0 {
				return nil
			}

			errPb.Trace.Operation = ctx.Route().Path
			code := errutil.GrpcCodeToHTTP(codes.Code(errPb.Code.StatusCode))
			ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			return ctx.Status(code).JSON(errPb.Code)
		},
	})

	if conf.EnableCors {
		httpServer.Use(cors.New(cors.Config{
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
	}

	httpApp := fiber.New()
	//app.Use(handlerHttpMiddle(globalMiddlewares))
	for _, h := range httpRouters {
		//srv := doc.WithService()
		//for _, an := range h.Annotation() {
		//	switch a := an.(type) {
		//	case *annotation.Openapi:
		//		if a.ServiceName != "" {
		//			srv.SetName(a.ServiceName)
		//		}
		//	}
		//}

		assert.If(h.Prefix() == "", "http handler prefix required")

		g := httpApp.Group(h.Prefix(), handlerHttpMiddle(append(globalMiddlewares, h.Middlewares()...)))
		h.Router(g)

		if m, ok := h.(lava.Close); ok {
			lifecycle.BeforeStop(m.Close)
		}

		if m, ok := h.(lava.Init); ok {
			s.initList = append(s.initList, m.Init)
		}
	}

	mux := gateway.NewMux()
	if len(gw) > 0 {
		mux = gw[0]
	}

	srvMidMap := make(map[string][]lava.Middleware)
	for _, h := range grpcRouters {
		desc := h.ServiceDesc()
		assert.If(desc == nil, "desc is nil")

		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], globalMiddlewares...)
		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], h.Middlewares()...)

		if m, ok := h.(lava.Close); ok {
			lifecycle.BeforeStop(m.Close)
		}

		if m, ok := h.(lava.Initializer); ok {
			s.initList = append(s.initList, m.Initialize)
		}

		if m, ok := h.(lava.Init); ok {
			s.initList = append(s.initList, m.Init)
		}

		mux.RegisterService(desc, h)
		s.cc.RegisterService(desc, h)
	}

	for _, h := range grpcProxy {
		desc := h.ServiceDesc()
		assert.If(desc == nil, "desc is nil")

		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], globalMiddlewares...)
		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], h.Middlewares()...)

		if m, ok := h.(lava.Close); ok {
			lifecycle.BeforeStop(m.Close)
		}

		if m, ok := h.(lava.Initializer); ok {
			s.initList = append(s.initList, m.Initialize)
		}

		if m, ok := h.(lava.Init); ok {
			s.initList = append(s.initList, m.Init)
		}

		cli := grpcc.New(
			&grpcc_config.Cfg{
				Service: &grpcc_config.ServiceCfg{
					Name:   h.Proxy().Name,
					Addr:   h.Proxy().Addr,
					Scheme: h.Proxy().Resolver,
				},
			},
			grpcc.Params{
				Log:    log,
				Metric: metric,
			},
			srvMidMap[desc.ServiceName]...,
		)

		mux.RegisterProxy(desc, h, cli)
	}

	mux.SetUnaryInterceptor(handlerUnaryMiddle(srvMidMap))
	mux.SetStreamInterceptor(handlerStreamMiddle(srvMidMap))
	s.cc = s.cc.WithServerUnaryInterceptor(handlerUnaryMiddle(srvMidMap))
	s.cc = s.cc.WithServerStreamInterceptor(handlerStreamMiddle(srvMidMap))

	// grpc server初始化
	grpcServer := conf.GrpcConfig.Build(
		grpc.ChainUnaryInterceptor(handlerUnaryMiddle(srvMidMap)),
		grpc.ChainStreamInterceptor(handlerStreamMiddle(srvMidMap)),
	).Expect("failed to build grpc server")

	for _, h := range grpcRouters {
		grpcServer.RegisterService(h.ServiceDesc(), h)
	}

	grpcGatewayApiPrefix := assert.Must1(url.JoinPath(conf.BaseUrl, "api"))
	s.log.Info().Str("path", grpcGatewayApiPrefix).Msg("service grpc gateway base path")

	for _, m := range mux.GetRouteMethods() {
		log.Info().
			Str("operation", m.Operation).
			Any("rpc-meta", mux.GetOperation(m.Operation).Meta).
			Str("http-method", m.Method).
			Str("http-path", "/"+strings.Trim(grpcGatewayApiPrefix, "/")+m.Path).
			Str("verb", m.Verb).
			Any("path-vars", m.Vars).
			Str("extras", fmt.Sprintf("%v", m.Extras)).
			Msg("grpc gateway router info")
	}

	httpServer.Mount("/debug", debug.App())
	httpServer.Mount(conf.BaseUrl, httpApp)
	httpServer.Group(grpcGatewayApiPrefix, httputil.StripPrefix(grpcGatewayApiPrefix, mux.Handler))

	s.httpServer = httpServer
	s.grpcServer = grpcServer

	vars.RegisterValue(fmt.Sprintf("%s-grpc-server-config-%s", version.Project(), xid.New()), &conf)
	vars.Register(fmt.Sprintf("%s-grpc-server-router-%s", version.Project(), xid.New()), func() interface{} {
		return mux.GetRouteMethods()
	})
	vars.Register(fmt.Sprintf("%s-grpc-server-desc-%s", version.Project(), xid.New()), func() interface{} {
		return grpcServer.GetServiceInfo()
	})
	vars.Register(fmt.Sprintf("%s-http-server-router-%s", version.Project(), xid.New()), func() interface{} {
		return httpServer.Stack()
	})
}

func (s *serviceImpl) start(ctx context.Context) (gErr error) {
	defer recovery.Exit()

	logutil.OkOrFailed(s.log, "running before service starts", func() error {
		for _, run := range s.lc.GetBeforeStarts() {
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Exec))
			assert.Exit(run.Exec(ctx))
		}
		return nil
	})

	logutil.OkOrFailed(s.log, "init handler before service starts", func() error {
		defer recovery.Exit()
		for _, init := range s.initList {
			s.log.Info().Msgf("init handler %s", stack.CallerWithFunc(init))
			init()
		}
		return nil
	})

	s.log.Info().
		Int("grpc-port", *s.conf.GrpcPort).
		Int("http-port", *s.conf.HttpPort).
		Msg("create network listener")
	grpcLn := assert.Exit1(net.Listen("tcp", fmt.Sprintf(":%d", *s.conf.GrpcPort)))
	httpLn := assert.Exit1(net.Listen("tcp", fmt.Sprintf(":%d", *s.conf.HttpPort)))

	logutil.OkOrFailed(s.log, "service starts", func() error {
		// 启动grpc服务
		async.GoDelay(func() error {
			s.log.Info().Msg("[grpc] Server Starting")
			logutil.LogOrErr(s.log, "[grpc] Server Stop", func() error {
				defer recovery.DebugPrint()
				err := s.grpcServer.Serve(grpcLn)
				if err == nil || errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed) {
					return nil
				}

				return err
			})
			return nil
		})

		// 启动grpc网关
		async.GoDelay(func() error {
			s.log.Info().Msg("[http] Server Starting")
			logutil.LogOrErr(s.log, "[http] Server Stop", func() error {
				defer recovery.DebugPrint()
				err := s.httpServer.Listener(httpLn)
				if err == nil || errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed) {
					return nil
				}

				return err
			})
			return nil
		})
		return nil
	})

	logutil.OkOrFailed(s.log, "running after service starts", func() error {
		for _, run := range s.lc.GetAfterStarts() {
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Exec))
			assert.Exit(run.Exec(ctx))
		}
		return nil
	})
	return nil
}

func (s *serviceImpl) stop(ctx context.Context) {
	defer recovery.DebugPrint()

	logutil.OkOrFailed(s.log, "running before service stops", func() error {
		for _, run := range s.lc.GetBeforeStops() {
			logutil.LogOrErr(
				s.log,
				fmt.Sprintf("running %s", stack.CallerWithFunc(run.Exec)),
				func() error { return run.Exec(ctx) },
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
				fmt.Sprintf("running %s", stack.CallerWithFunc(run.Exec)),
				func() error { return run.Exec(ctx) },
			)
		}
		return nil
	})
}
