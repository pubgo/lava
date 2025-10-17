package grpcs

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/async"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/stack"
	"github.com/pubgo/funk/v2/vars"
	"github.com/rs/xid"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/v2/clients/grpcc"
	"github.com/pubgo/lava/v2/clients/grpcc/grpccconfig"
	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/metrics"
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
	GrpcProxy       []lava.GrpcProxy
	DixMiddlewares  []lava.Middleware
	Metric          metrics.Metric
	Log             log.Logger
	Conf            *Config
	Gw              []*gateway.Mux
}

func New(params Params) supervisor.Service { return newService(params) }

func newService(params Params) supervisor.Service {
	s := &serviceImpl{cc: new(inprocgrpc.Channel)}
	s.init(
		params.GrpcRouters,
		params.HttpRouters,
		params.GrpcHttpRouters,
		params.GrpcProxy,
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
	log        log.Logger
	cc         *inprocgrpc.Channel
	initList   []func()
	conf       *Config
}

func (s *serviceImpl) String() string {
	return "grpc-server"
}

func (s *serviceImpl) Serve(ctx context.Context) (err error) {
	defer s.stop(ctx)
	err = s.start(ctx)
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
	grpcProxy []lava.GrpcProxy,
	dixMiddlewares []lava.Middleware,
	metric metrics.Metric,
	log log.Logger,
	conf *Config,
	gw []*gateway.Mux,
) {
	cfg := httputil.DefaultCfg(&httputil.Config{
		BaseUrl:           conf.BaseUrl,
		EnablePrintRouter: conf.EnablePrintRouter,
		Http:              conf.Http,
		HttpPort:          conf.HttpPort,
	})
	conf.BaseUrl = cfg.BaseUrl
	conf.HttpPort = cfg.HttpPort
	conf.Http = cfg.Http

	s.conf = config.MergeR(defaultCfg(), conf).Must()

	globalMiddlewares := lava.Middlewares{
		middleware_serviceinfo.New(),
		middleware_metric.New(metric),
		middleware_accesslog.New(log),
		middleware_recovery.New(),
	}
	globalMiddlewares = append(globalMiddlewares, dixMiddlewares...)

	log = log.WithName("grpc-server")
	s.log = log

	httpServer := fiber.New(conf.Http.Build().Must())
	httpServer.Use(httputil.Cors())

	httpApp := fiber.New()

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

	for _, h := range httpRouters {
		assert.If(h.Prefix() == "", "http handler prefix required")

		g := httpApp.Group(h.Prefix(), handlerHttpMiddle(append(globalMiddlewares, h.Middlewares()...)))
		h.Router(g)
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

		mux.RegisterService(desc, h)
		s.cc.RegisterService(desc, h)
	}

	for _, h := range grpcProxy {
		desc := h.ServiceDesc()
		assert.If(desc == nil, "desc is nil")

		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], globalMiddlewares...)
		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], h.Middlewares()...)

		cli := grpcc.New(
			&grpccconfig.Cfg{
				Service: &grpccconfig.ServiceCfg{
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
			Str("verb", m.Verb).
			Any("path-vars", m.Vars).
			Str("extras", fmt.Sprintf("%v", m.Extras)).
			Msgf("grpc gateway router info: %s %s", m.Method, "/"+strings.Trim(grpcGatewayApiPrefix, "/")+m.Path)
	}

	httpServer.Mount("/debug", debug.App())
	httpServer.Mount(conf.BaseUrl, httpApp)
	httpServer.Group(grpcGatewayApiPrefix, httputil.StripPrefix(grpcGatewayApiPrefix, mux.Handler))

	s.httpServer = httpServer
	s.grpcServer = grpcServer

	vars.Register(fmt.Sprintf("%s-grpc-server-config-%s", version.Project(), xid.New()), func() any { return conf })
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

	// 启动grpc服务
	async.GoDelay(func() error {
		s.log.Info().Msg("[grpc] Server Starting")
		defer recovery.DebugPrint()
		err := s.grpcServer.Serve(grpcLn)
		if netutil.IsErrServerClosed(err) {
			return nil
		}

		return err
	})

	// 启动grpc网关
	async.GoDelay(func() error {
		s.log.Info().Msg("[http] Server Starting")
		defer recovery.DebugPrint()
		err := s.httpServer.Listener(httpLn)
		if netutil.IsErrServerClosed(err) {
			return nil
		}

		return err
	})

	return nil
}

func (s *serviceImpl) stop(ctx context.Context) {
	defer recovery.DebugPrint()

	logutil.LogOrErr(s.log, "[grpc] Server GracefulStop", func() error {
		s.grpcServer.GracefulStop()
		return nil
	})

	logutil.LogOrErr(s.log, "[http] Server Shutdown", func() error {
		err := s.httpServer.ShutdownWithContext(ctx)
		if netutil.IsErrServerClosed(err) {
			return nil
		}
		return err
	})
}
