package grpcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	fiber "github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
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
	"github.com/pubgo/funk/typex"
	"github.com/pubgo/funk/vars"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
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
	"github.com/pubgo/lava/pkg/gateway"
	"github.com/pubgo/lava/pkg/httputil"
	"github.com/pubgo/lava/pkg/wsproxy"
	"github.com/rs/xid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func New() lava.Service { return newService() }

func newService() *serviceImpl {
	return &serviceImpl{
		cc: new(inprocgrpc.Channel),
	}
}

var _ lava.Service = (*serviceImpl)(nil)

type serviceImpl struct {
	lc         lifecycle.Getter
	httpServer *fiber.App
	grpcServer *grpc.Server
	log        log.Logger
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
		AppName:            version.Project(),
		BodyLimit:          100 * 1024 * 1024,
		ErrorHandler: func(ctx fiber.Ctx, err error) error {
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
			AllowMethods: []string{
				fiber.MethodGet,
				fiber.MethodPost,
				fiber.MethodPut,
				fiber.MethodDelete,
				fiber.MethodPatch,
				fiber.MethodHead,
				fiber.MethodOptions,
			},
			//AllowHeaders:     "",
			AllowCredentials: true,
			//ExposeHeaders:    "",
			MaxAge: 0,
		}))
	}

	app := fiber.New()
	app.Group("/debug", httputil.StripPrefix(filepath.Join(conf.BaseUrl, "/debug"), debug.Handler))

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

		if h.Prefix() == "" {
			panic("http handler prefix is required")
		}

		g := app.Group(h.Prefix(), handlerHttpMiddle(append(globalMiddlewares, h.Middlewares()...)))
		h.Router(g)

		if m, ok := h.(lava.Close); ok {
			lifecycle.BeforeStop(m.Close)
		}

		if m, ok := h.(lava.Init); ok {
			s.initList = append(s.initList, m.Init)
		}
	}

	for _, handler := range grpcRouters {
		h, ok := handler.(lava.HttpRouter)
		if !ok {
			continue
		}

		if h.Prefix() == "" {
			panic("http handler prefix is required")
		}

		g := app.Group(h.Prefix(), handlerHttpMiddle(append(globalMiddlewares, h.Middlewares()...)))
		h.Router(g)

		if m, ok := h.(lava.Close); ok {
			lifecycle.BeforeStop(m.Close)
		}

		if m, ok := h.(lava.Init); ok {
			s.initList = append(s.initList, m.Init)
		}
	}

	httpServer.Use(conf.BaseUrl, app)

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
					Code:       int32(errorpb.Code_Internal),
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
						Code:       int32(errorpb.Code(sts.Code())),
						StatusCode: errorpb.Code(sts.Code()),
						Name:       "lava.grpc.status",
						Details:    sts.Proto().Details,
					}
				}
			}

			const fallback = `{"code":13, "name":"lava.grpc.status", "status_code": 500, "message": "failed to marshal error message"}`

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
		if m, ok := h.(lava.GrpcGatewayRouter); ok {
			assert.Exit(m.RegisterGateway(context.Background(), grpcGateway, s.cc))
		}
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

	typex.DoBlock(func() {
		apiPrefix1 := assert.Must1(url.JoinPath(conf.BaseUrl, "gw"))
		s.log.Info().Str("path", apiPrefix1).Msg("service grpc gateway base path")
		httpServer.Group(apiPrefix1, httputil.StripPrefix(apiPrefix1, mux.Handler))
		for _, m := range mux.GetRouteMethods() {
			log.Info().
				Str("operation", m.Operation).
				Any("rpc-meta", mux.GetOperation(m.Operation).Meta).
				Str("http-method", m.Method).
				Str("http-path", "/"+strings.Trim(apiPrefix1, "/")+m.Path).
				Str("verb", m.Verb).
				Any("path-vars", m.Vars).
				Str("extras", fmt.Sprintf("%v", m.Extras)).
				Msg("grpc gateway router info")
		}
	})

	typex.DoBlock(func() {
		apiPrefix := assert.Must1(url.JoinPath(conf.BaseUrl, "api"))
		s.log.Info().Str("path", apiPrefix).Msg("service grpc gateway base path")
		httpServer.Group(apiPrefix, httputil.HTTPHandler(http.StripPrefix(apiPrefix, wsproxy.WebsocketProxy(grpcGateway,
			wsproxy.WithPingPong(conf.EnablePingPong),
			wsproxy.WithTimeWait(conf.PingPongTime),
			wsproxy.WithReadLimit(int64(generic.FromPtr(conf.WsReadLimit))),
		))))
	})

	s.httpServer = httpServer
	s.grpcServer = grpcServer

	vars.RegisterValue(fmt.Sprintf("%s-grpc-server-config-%s", version.Project(), xid.New()), &conf)
	// vars.RegisterValue(fmt.Sprintf("%s-grpc-server-router-%s", version.Project(), xid.New()), mux.App().Stack())
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
