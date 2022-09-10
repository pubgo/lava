package grpcs

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/pubgo/lava/debug"
	"github.com/twitchtv/twirp"
	"net"
	"net/http"
	"strings"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/logx"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/syncx"
	"github.com/pubgo/x/stack"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/logging/logutil"
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
	handlers []service.GrpcHandler,
	middlewares []service.Middleware,
	getLifecycle lifecycle.GetLifecycle,
	lifecycle lifecycle.Lifecycle,
	log *zap.Logger,
	cfg *Cfg) {

	log = log.Named("grpc-server")

	s.lc = getLifecycle
	s.log = log

	var httpServer = cfg.Api.Build().Unwrap()
	s.httpServer = httpServer

	httpServer.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowCredentials: true,
	}))

	app := chi.NewRouter()
	for _, h := range handlers {
		s.initList = append(s.initList, h.Init)

		middlewares = append(middlewares, h.Middlewares()...)
		var hh = h.TwirpHandler(twirp.WithServerPathPrefix(cfg.BasePrefix))
		app.Mount(cfg.BasePrefix+h.ServiceDesc().ServiceName, hh)

		if m, ok := h.(service.Close); ok {
			lifecycle.BeforeStop(m.Close)
		}
	}

	httpServer.Use(handlerHttpMiddle(middlewares))
	httpServer.Mount("/debug", debug.App())

	// grpc server初始化
	var grpcServer = cfg.Grpc.Build(
		grpc.ChainUnaryInterceptor(handlerUnaryMiddle(middlewares)),
		grpc.ChainStreamInterceptor(handlerStreamMiddle(middlewares)),
	).Unwrap()
	s.grpcServer = grpcServer

	wrappedGrpc := grpcweb.WrapServer(grpcServer,
		grpcweb.WithAllowNonRootResource(true),
		grpcweb.WithOriginFunc(func(origin string) bool { return true }))

	httpServer.All(cfg.BasePrefix, adaptor.HTTPHandler(h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, stripPrefix(cfg.BasePrefix, r))
			return
		}

		if wrappedGrpc.IsGrpcWebSocketRequest(r) {
			wrappedGrpc.HandleGrpcWebsocketRequest(w, stripPrefix(cfg.BasePrefix, r))
			return
		}

		if wrappedGrpc.IsGrpcWebRequest(r) {
			wrappedGrpc.HandleGrpcWebRequest(w, stripPrefix(cfg.BasePrefix, r))
			return
		}

		app.ServeHTTP(w, r)
	}), &http2.Server{})))

	// 网关初始化

	if cfg.PrintRoute {
		for _, stacks := range httpServer.Stack() {
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

func stripPrefix(prefix string, r *http.Request) *http.Request {
	p := strings.TrimPrefix(r.URL.Path, prefix)
	rp := strings.TrimPrefix(r.URL.RawPath, prefix)
	if len(p) < len(r.URL.Path) && (r.URL.RawPath == "" || len(rp) < len(r.URL.RawPath)) {
		r.URL.Path = p
		r.URL.RawPath = rp
	}
	return r
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

	grpcLn := assert.Must1(net.Listen("tcp", fmt.Sprintf(":%d", runmode.GrpcPort)))
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
		// 启动grpc服务
		syncx.GoDelay(func() result.Error {
			s.log.Info("[grpc] Server Starting")
			logutil.LogOrErr(s.log, "[grpc] Server Stop", func() result.Error {
				defer recovery.Exit()
				if err := s.grpcServer.Serve(grpcLn); err != nil &&
					!errors.Is(err, http.ErrServerClosed) &&
					!errors.Is(err, net.ErrClosed) {
					return result.WithErr(err)
				}
				return result.NilErr()
			})
			return result.NilErr()
		})

		// 启动grpc网关
		syncx.GoDelay(func() result.Error {
			s.log.Info("[grpc-gw] Server Starting")
			logutil.LogOrErr(s.log, "[grpc-gw] Server Stop", func() result.Error {
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

	logutil.LogOrErr(s.log, "[grpc-gateway] Shutdown", func() result.Error {
		return result.WithErr(s.httpServer.Shutdown())
	})

	logutil.LogOrErr(s.log, "[grpc] GracefulStop", func() result.Error {
		s.grpcServer.GracefulStop()
		return result.NilErr()
	})

	logutil.OkOrFailed(s.log, "service after-stop", func() result.Error {
		for _, run := range s.lc.GetAfterStops() {
			s.log.Sugar().Infof("running %s", stack.Func(run.Handler))
			run.Handler()
		}
		return result.NilErr()
	})
}
