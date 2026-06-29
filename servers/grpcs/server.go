package grpcs

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/async"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/vars"
	"github.com/samber/lo"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/core/running"
	"github.com/pubgo/lava/v2/core/supervisor"
	"github.com/pubgo/lava/v2/internal/logutil"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_serviceinfo"
	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/gateway"
	"github.com/pubgo/lava/v2/pkg/httputil"
	"github.com/pubgo/lava/v2/pkg/netutil"
)

type Params struct {
	GrpcRouters     []lava.GrpcRouter
	HttpRouters     []lava.HttpRouter
	GrpcHttpRouters []lava.GrpcHttpRouter
	// GrpcProxy       []lava.GrpcProxy
	DixMiddlewares []lava.Middleware
	Metric         metrics.Metric
	Log            log.Logger
	Conf           *Config
	Gw             []*gateway.Mux
}

func New(params Params) supervisor.Service { return newService(params) }

func newService(params Params) supervisor.Service {
	s := &serviceImpl{}
	s.init(
		params.GrpcRouters,
		params.HttpRouters,
		params.GrpcHttpRouters,
		// params.GrpcProxy,
		params.DixMiddlewares,
		params.Metric,
		params.Log,
		params.Conf,
		params.Gw,
	)

	return supervisor.NewService("grpc-server", s.Serve)
}

type serviceImpl struct {
	httpServer *fiber.App
	grpcServer *grpc.Server
	wsServer   *http.Server
	log        log.Logger
	conf       *Config
}

func (s *serviceImpl) String() string {
	return "grpc-server"
}

func (s *serviceImpl) Serve(ctx context.Context) error {
	defer s.stop(ctx)
	err := s.start(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func (s *serviceImpl) init(
	grpcRouters []lava.GrpcRouter,
	httpRouters []lava.HttpRouter,
	grpcHttpRouters []lava.GrpcHttpRouter,
	// grpcProxy []lava.GrpcProxy,
	dixMiddlewares []lava.Middleware,
	metric metrics.Metric,
	log log.Logger,
	conf *Config,
	gw []*gateway.Mux,
) {
	cfg := httputil.DefaultCfg(&httputil.Config{
		EnablePrintRouter: conf.EnablePrintRouter,
		Http:              conf.Http,
	})
	conf.Http = cfg.Http

	s.conf = config.MergeR(defaultCfg(), conf).Unwrap()

	globalMiddlewares := lava.Middlewares{
		middleware_serviceinfo.New(),
		middleware_metric.New(metric),
		middleware_accesslog.New(log),
	}
	globalMiddlewares = append(globalMiddlewares, dixMiddlewares...)
	globalMiddlewares = append(globalMiddlewares, middleware_recovery.New())

	log = log.WithName("grpc-server")

	httpServer := fiber.New(conf.Http.Build().Unwrap())
	httpServer.Use(httputil.Cors())
	httpServer.Use(func(ctx fiber.Ctx) error {
		log.Debug().
			Str("path", ctx.Path()).
			Str("method", ctx.Method()).
			Str("header", ctx.Request().Header.String()).
			Msg("grpc gateway router")
		return ctx.Next()
	})

	for _, h := range grpcRouters {
		r, ok := h.(lava.HttpRouter)
		if !ok {
			continue
		}

		httpRouters = append(httpRouters, r)
	}

	for _, h := range grpcHttpRouters {
		if r, ok := h.(lava.HttpRouter); ok {
			httpRouters = append(httpRouters, r)
		}

		if r, ok := h.(lava.GrpcRouter); ok {
			grpcRouters = append(grpcRouters, r)
		}
	}

	httpApp := fiber.New()
	for _, h := range httpRouters {
		assert.If(h.Prefix() == "", "http router prefix required")

		h.Router(httpApp.Group(h.Prefix(), handlerHttpMiddle(append(globalMiddlewares, h.Middlewares()...))))
	}

	mux := gateway.NewMux()
	if len(gw) > 0 {
		mux = gw[0]
	}

	srvMidMap := make(map[string][]lava.Middleware)
	for _, h := range grpcRouters {
		desc := h.ServiceDesc()
		assert.If(desc == nil, "service desc is nil")

		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], globalMiddlewares...)
		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], h.Middlewares()...)

		mux.RegisterService(desc, h)
	}

	//for _, h := range grpcProxy {
	//	desc := h.ServiceDesc()
	//	assert.If(desc == nil, "service desc is nil")
	//
	//	srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], globalMiddlewares...)
	//	srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], h.Middlewares()...)
	//
	//	cli := grpcc.New(
	//		&grpccconfig.Cfg{
	//			Service: &grpccconfig.ServiceCfg{
	//				Name:   h.Proxy().Name,
	//				Addr:   h.Proxy().Addr,
	//				Scheme: h.Proxy().Resolver,
	//			},
	//		},
	//		grpcc.Params{
	//			Log:    log,
	//			Metric: metric,
	//		},
	//		srvMidMap[desc.ServiceName]...,
	//	)
	//
	//	mux.RegisterProxy(desc, h, cli)
	//}

	mux.SetUnaryInterceptor(handlerUnaryMiddle(srvMidMap))
	mux.SetStreamInterceptor(handlerStreamMiddle(srvMidMap))

	grpcServerOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(handlerUnaryMiddle(srvMidMap)),
		grpc.ChainStreamInterceptor(handlerStreamMiddle(srvMidMap)),
	}
	if conf.GRPCPassthrough {
		grpcServerOpts = mux.GRPCServerOptions(grpcServerOpts...)
		log.Info().Msg("gateway grpc passthrough enabled: register services on Mux only")
	}

	// grpc server初始化
	grpcServer := conf.GrpcConfig.Build(grpcServerOpts...).Expect("failed to build grpc server")

	if !conf.GRPCPassthrough {
		for _, h := range grpcRouters {
			grpcServer.RegisterService(h.ServiceDesc(), h)
		}

		for _, h := range grpcHttpRouters {
			grpcServer.RegisterService(h.ServiceDesc(), h)
		}
	}

	//for _, h := range grpcProxy {
	//	grpcServer.RegisterService(h.ServiceDesc(), h)
	//}

	grpcGatewayApiPrefix := "/api"
	log.Info().Msgf("service gateway base path: %s", grpcGatewayApiPrefix)

	if conf.EnablePrintRouter {
		for _, m := range mux.GetRouteMethods() {
			log.Info().
				Str("operation", m.Operation).
				Any("rpc-meta", lo.FromPtr(mux.GetOperation(m.Operation)).Meta).
				Str("verb", m.Verb).
				Any("path-vars", m.Vars).
				Str("extras", fmt.Sprintf("%v", m.Extras)).
				Msgf("grpc gateway router info: %s %s", m.Method, "/"+strings.Trim(grpcGatewayApiPrefix, "/")+m.Path)
		}
	}

	httpServer.Use("/debug", debug.App())
	httpServer.Use("/", httpApp)
	httpServer.Use(grpcGatewayApiPrefix, func(ctx fiber.Ctx) error {
		return httputil.StripPrefix(grpcGatewayApiPrefix, mux.Handler)(ctx)
	})

	if conf.WebSocketPort > 0 {
		wsOpts := gateway.WSOptionsFromConfig(gateway.WSConfig{
			OriginPatterns:     conf.WebSocketOriginPatterns,
			InsecureSkipVerify: conf.WebSocketInsecureSkipVerify || len(conf.WebSocketOriginPatterns) == 0,
		})
		wsHandler := mux.WebSocketHandler(wsOpts...)
		s.wsServer = &http.Server{
			Addr:              fmt.Sprintf(":%d", conf.WebSocketPort),
			Handler:           wsHandler,
			ReadHeaderTimeout: 10 * time.Second,
		}
		log.Info().Int64("port", conf.WebSocketPort).Msg("gateway websocket server enabled")
	}

	s.log = log
	s.httpServer = httpServer
	s.grpcServer = grpcServer

	vars.Register(vars.UniqueName(version.Project(), "grpc-server-info"), func() any {
		return map[string]any{
			"config": conf,
			"method": mux.GetRouteMethods(),
			"desc":   grpcServer.GetServiceInfo(),
			"router": httpServer.Stack(),
		}
	})
}

func (s *serviceImpl) start(ctx context.Context) (gErr error) {
	defer recovery.Exit()

	s.log.Info().
		Int64("grpc-port", running.GrpcPort.Value()).
		Int64("http-port", running.HttpPort.Value()).
		Msg("create network listener")
	grpcLn := assert.Exit1(net.Listen("tcp", fmt.Sprintf(":%d", running.GrpcPort.Value())))
	httpLn := assert.Exit1(net.Listen("tcp", fmt.Sprintf(":%d", running.HttpPort.Value())))

	async.GoDelay(func() error {
		s.log.Info().Msg("grpc server starting")
		defer recovery.DebugPrint()
		err := s.grpcServer.Serve(grpcLn)
		if netutil.IsErrServerClosed(err) {
			return nil
		}

		return err
	})

	// 启动grpc网关
	async.GoDelay(func() error {
		s.log.Info().Msg("http server starting")
		defer recovery.DebugPrint()
		err := s.httpServer.Listener(httpLn)
		if netutil.IsErrServerClosed(err) {
			return nil
		}

		return err
	})

	if s.wsServer != nil {
		wsLn := assert.Exit1(net.Listen("tcp", s.wsServer.Addr))
		async.GoDelay(func() error {
			s.log.Info().Str("addr", s.wsServer.Addr).Msg("websocket server starting")
			defer recovery.DebugPrint()
			err := s.wsServer.Serve(wsLn)
			if netutil.IsErrServerClosed(err) {
				return nil
			}
			return err
		})
	}

	return nil
}

func (s *serviceImpl) stop(ctx context.Context) {
	defer recovery.DebugPrint()

	logutil.LogOrErr(s.log, "grpc server graceful stop", func() error {
		s.grpcServer.GracefulStop()
		return nil
	})

	logutil.LogOrErr(s.log, "http server shutdown", func() error {
		err := s.httpServer.ShutdownWithContext(ctx)
		if netutil.IsErrServerClosed(err) {
			return nil
		}
		return err
	})

	if s.wsServer != nil {
		logutil.LogOrErr(s.log, "websocket server shutdown", func() error {
			err := s.wsServer.Shutdown(ctx)
			if netutil.IsErrServerClosed(err) {
				return nil
			}
			return err
		})
	}
}
