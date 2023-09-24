package grpcs

import (
	"context"
	"errors"
	"fmt"
	"github.com/pubgo/funk/proto/errorpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/fullstorydev/grpchan"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/funk/vars"
	"github.com/pubgo/funk/version"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

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
	"github.com/pubgo/lava/pkg/grpcutil"
)

func New() lava.Service { return newService() }

func newService() *serviceImpl {
	return &serviceImpl{
		handlers: make(grpchan.HandlerMap),
		cc:       new(inprocgrpc.Channel),
	}
}

var _ lava.Service = (*serviceImpl)(nil)

type serviceImpl struct {
	lc         lifecycle.Getter
	httpServer *fiber.App
	grpcServer *grpc.Server
	log        log.Logger
	handlers   grpchan.HandlerMap
	cc         *inprocgrpc.Channel
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
	handlers []lava.GrpcRouter,
	dixMiddlewares map[string][]lava.Middleware,
	getLifecycle lifecycle.Getter,
	lifecycle lifecycle.Lifecycle,
	metric metrics.Metric,
	log log.Logger,
	conf *Config,
) {
	s.lc = getLifecycle

	conf = config.MergeR(defaultCfg(), conf).Unwrap()
	basePath := "/" + strings.Trim(conf.BaseUrl, "/")
	conf.BaseUrl = basePath

	middlewares := lava.Middlewares{
		middleware_service_info.New(),
		middleware_metric.New(metric),
		middleware_accesslog.New(log),
		middleware_recovery.New(),
	}

	// TODO server middleware handle
	if dixMiddlewares != nil {
		middlewares = append(middlewares, dixMiddlewares["server"]...)
	}

	log = log.WithName("grpc-server")
	s.log = log

	httpServer := fiber.New(fiber.Config{
		EnableIPValidation: true,
		EnablePrintRoutes:  conf.EnablePrintRoutes,
		AppName:            version.Project(),
	})
	httpServer.Mount("/debug", debug.App())

	apiPrefix := assert.Must1(url.JoinPath(basePath, "api"))

	grpcGateway := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
			return strings.ToLower(s), true
		}),
		runtime.WithOutgoingHeaderMatcher(func(s string) (string, bool) {
			return strings.ToUpper(s), true
		}),
		runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			path, ok := runtime.HTTPPathPattern(ctx)
			if !ok {
				return nil
			}
			return metadata.Pairs("http_path", path)
		}),
		runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, request *http.Request, err error) {
			sts, ok := status.FromError(err)
			if !ok {
				runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, request, err)
				return
			}

			const fallback = `{"code": 13, "message": "failed to marshal error message"}`

			errpb := sts.Details()[0].(*errorpb.Error)

			s := status.Convert(err)
			pb := s.Proto()

			w.Header().Del("Trailer")
			w.Header().Del("Transfer-Encoding")

			w.Header().Set("Content-Type", marshaler.ContentType(pb))

			if s.Code() == codes.Unauthenticated {
				w.Header().Set("WWW-Authenticate", s.Message())
			}

			buf, merr := marshaler.Marshal(pb)
			if merr != nil {
				grpclog.Infof("Failed to marshal error message %q: %v", s, merr)
				w.WriteHeader(http.StatusInternalServerError)
				if _, err := io.WriteString(w, fallback); err != nil {
					grpclog.Infof("Failed to write response: %v", err)
				}
				return
			}

			md, ok := runtime.ServerMetadataFromContext(ctx)
			if ok && w != nil {
				for k, v := range md.HeaderMD {
					for i := range v {
						w.Header().Add(k, v[i])
					}
				}

				for k, v := range md.TrailerMD {
					for i := range v {
						w.Header().Add(k, v[i])
					}
				}
			}

			w.WriteHeader(runtime.HTTPStatusFromCode(s.Code()))
			if _, err := w.Write(buf); err != nil {
				grpclog.Infof("Failed to write response: %v", err)
			}
		}),
	)

	srvMidMap := make(map[string][]lava.Middleware)
	for _, h := range handlers {
		desc := h.ServiceDesc()
		assert.If(desc == nil, "desc is nil")

		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], middlewares...)
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

		s.cc.RegisterService(desc, h)
		if m, ok := h.(lava.GrpcGatewayRouter); ok {
			assert.Exit(m.RegisterGateway(context.Background(), grpcGateway, s.cc))
		}
	}

	s.cc = s.cc.WithServerUnaryInterceptor(handlerUnaryMiddle(srvMidMap))
	s.cc = s.cc.WithServerStreamInterceptor(handlerStreamMiddle(srvMidMap))

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

	s.log.Info().Str("path", apiPrefix).Msg("service web base path")
	httpServer.Group(apiPrefix+"/*", adaptor.HTTPHandler(h2c.NewHandler(http.StripPrefix(apiPrefix, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
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

		if grpcutil.IsGRPCRequest(request) {
			grpcServer.ServeHTTP(writer, request)
			return
		}

		grpcGateway.ServeHTTP(writer, request)
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
		Int("grpc-port", running.GrpcPort).
		Int("http-port", running.HttpPort).
		Msg("create network listener")
	grpcLn := assert.Must1(net.Listen("tcp", fmt.Sprintf(":%d", running.GrpcPort)))
	httpLn := assert.Must1(net.Listen("tcp", fmt.Sprintf(":%d", running.HttpPort)))

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
