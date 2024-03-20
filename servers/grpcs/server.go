package grpcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/fullstorydev/grpchan"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/errors/errutil"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/proto/errorpb"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/funk/vars"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/pkg/gateway"
	"github.com/pubgo/lava/pkg/httputil"
	"github.com/pubgo/lava/pkg/wsproxy"
	"github.com/pubgo/opendoc/opendoc"
	"github.com/rs/xid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

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
	return &serviceImpl{
		reg: make(grpchan.HandlerMap),
		cc:  new(inprocgrpc.Channel),
	}
}

var _ lava.Service = (*serviceImpl)(nil)

type serviceImpl struct {
	lc         lifecycle.Getter
	httpServer *fiber.App
	grpcServer *grpc.Server
	log        log.Logger
	reg        grpchan.HandlerMap
	cc         *inprocgrpc.Channel
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
	gw []*gateway.Mux,
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
		BodyLimit:          100 * 1024 * 1024,
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
		//srv := doc.WithService()
		//for _, an := range h.Annotation() {
		//	switch a := an.(type) {
		//	case *annotation.Openapi:
		//		if a.ServiceName != "" {
		//			srv.SetName(a.ServiceName)
		//		}
		//	}
		//}

		if h.Prefix() == "" {
			panic("http handler prefix is required")
		}

		var g = app.Group(h.Prefix(), handlerHttpMiddle(h.Middlewares()))
		h.Router(g)

		if m, ok := h.(lava.Close); ok {
			lifecycle.BeforeStop(m.Close)
		}

		if m, ok := h.(lava.Init); ok {
			s.initList = append(s.initList, m.Init)
		}
	}

	for _, handler := range handlers {
		//srv := doc.WithService()
		//for _, an := range h.Annotation() {
		//	switch a := an.(type) {
		//	case *annotation.Openapi:
		//		if a.ServiceName != "" {
		//			srv.SetName(a.ServiceName)
		//		}
		//	}
		//}

		h, ok := handler.(lava.HttpRouter)
		if !ok {
			continue
		}

		if h.Prefix() == "" {
			panic("http handler prefix is required")
		}

		var g = app.Group(h.Prefix(), handlerHttpMiddle(h.Middlewares()))
		h.Router(g)

		if m, ok := h.(lava.Close); ok {
			lifecycle.BeforeStop(m.Close)
		}

		if m, ok := h.(lava.Init); ok {
			s.initList = append(s.initList, m.Init)
		}
	}

	httpServer.Mount(conf.BaseUrl, app)

	grpcGateway := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
		runtime.SetQueryParameterParser(new(DefaultQueryParser)),
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
			return metadata.Pairs("http_path", path, "http_method", request.Method, "http_url", request.URL.Path)
		}),
		runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshal runtime.Marshaler, w http.ResponseWriter, request *http.Request, err error) {
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

			var pb *errorpb.ErrCode
			sts, ok := status.FromError(err)
			if !ok || sts == nil {
				w.Header().Set("Content-Type", "application/json")
				pb = &errorpb.ErrCode{
					Message:    err.Error(),
					StatusCode: errorpb.Code_Internal,
					Code:       500,
					Name:       "lava.grpc.status",
				}
			} else {
				w.Header().Set("Content-Type", marshal.ContentType(sts))
				if len(sts.Details()) > 0 {
					if code, ok := sts.Details()[0].(*errorpb.Error); ok {
						pb = code.Code
					}
				} else {
					pb = &errorpb.ErrCode{
						Message:    sts.Message(),
						StatusCode: errorpb.Code(sts.Code()),
						Name:       "lava.grpc.status",
						Details:    sts.Proto().Details,
					}
				}
			}

			const fallback = `{"code":500, "name":"lava.grpc.status", "status_code": 13, "message": "failed to marshal error message"}`

			// skip error
			if pb.StatusCode == errorpb.Code_OK {
				return
			}

			buf, mErr := marshal.Marshal(pb)
			if mErr != nil {
				grpclog.Infof("Failed to marshal error message %q: %v", pb, mErr)
				w.WriteHeader(http.StatusInternalServerError)
				if _, err := io.WriteString(w, fallback); err != nil {
					grpclog.Infof("Failed to write response: %v", err)
				}
				return
			}

			w.WriteHeader(runtime.HTTPStatusFromCode(codes.Code(pb.StatusCode)))
			if _, err := w.Write(buf); err != nil {
				grpclog.Infof("Failed to write response: %v", err)
			}
		}),
	)

	var mux = gateway.NewMux()
	if len(gw) > 0 {
		mux = gw[0]
	}
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

		mux.RegisterService(desc, h)
		s.reg.RegisterService(desc, h)
		s.cc.RegisterService(desc, h)
		if m, ok := h.(lava.GrpcGatewayRouter); ok {
			assert.Exit(m.RegisterGateway(context.Background(), grpcGateway, s.cc))
		}
	}

	mux.SetUnaryInterceptor(handlerUnaryMiddle(srvMidMap))
	mux.SetStreamInterceptor(handlerStreamMiddle(srvMidMap))
	s.cc = s.cc.WithServerUnaryInterceptor(handlerUnaryMiddle(srvMidMap))
	s.cc = s.cc.WithServerStreamInterceptor(handlerStreamMiddle(srvMidMap))

	// grpc server初始化
	grpcServer := conf.GrpcConfig.Build(
		grpc.ChainUnaryInterceptor(handlerUnaryMiddle(srvMidMap)),
		grpc.ChainStreamInterceptor(handlerStreamMiddle(srvMidMap))).Unwrap()

	for _, h := range handlers {
		grpcServer.RegisterService(h.ServiceDesc(), h)
	}

	apiPrefix1 := assert.Must1(url.JoinPath(conf.BaseUrl, "gw"))
	s.log.Info().Str("path", apiPrefix1).Msg("service grpc gateway base path")
	httpServer.Group(apiPrefix1, httputil.StripPrefix(apiPrefix1, mux.Handler))

	apiPrefix := assert.Must1(url.JoinPath(conf.BaseUrl, "api"))
	s.log.Info().Str("path", apiPrefix).Msg("service grpc gateway base path")
	httpServer.Group(apiPrefix, httputil.HTTPHandler(http.StripPrefix(apiPrefix, wsproxy.WebsocketProxy(grpcGateway,
		wsproxy.WithPingPong(conf.EnablePingPong),
		wsproxy.WithTimeWait(conf.PingPongTime),
	))))

	s.httpServer = httpServer
	s.grpcServer = grpcServer

	vars.RegisterValue(fmt.Sprintf("%s-grpc-server-config-%s", version.Project(), xid.New()), &conf)
	//vars.RegisterValue(fmt.Sprintf("%s-grpc-server-router-%s", version.Project(), xid.New()), mux.App().Stack())
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
