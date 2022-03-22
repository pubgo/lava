package service

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	fiber2 "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/soheilhy/cmux"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"github.com/pubgo/lava/config"
	encoding3 "github.com/pubgo/lava/encoding"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/service/internal/fiber_builder"
	"github.com/pubgo/lava/service/internal/grpc_builder"
	"github.com/pubgo/lava/service/service_type"
)

var _ service_type.Service = (*implService)(nil)

func newService(name string, desc string) *implService {
	var g = &implService{
		cmd:    &cli.Command{Name: name, Usage: desc},
		srv:    grpc_builder.New(),
		gw:     fiber_builder.New(),
		inproc: &inprocgrpc.Channel{},
		net: &netCfg{
			Addr:        "0.0.0.0",
			Port:        8080,
			ReadTimeout: time.Second * 2,
			HandleError: func(err error) bool {
				if errors.Is(err, net.ErrClosed) {
					return true
				}

				zap.L().Named("cmux").Error("cmux match failed", logutil.ErrField(err)...)
				return true
			},
		},
		cfg: Cfg{
			name:     name,
			hostname: runtime.Hostname,
			id:       uuid.New().String(),
			Grpc:     grpc_builder.GetDefaultCfg(),
			Gw:       fiber_builder.Cfg{},
		},
	}

	return g
}

type implService struct {
	beforeStarts []func()
	afterStarts  []func()
	beforeStops  []func()
	afterStops   []func()
	pluginList   []plugin.Plugin
	middlewares  []service_type.Middleware
	services     []service_type.Desc

	cmd *cli.Command

	L *logging.Logger `name:"service"`

	net *netCfg

	cfg Cfg
	srv grpc_builder.Builder
	gw  fiber_builder.Builder

	// inproc Channel is used to serve grpc gateway
	inproc *inprocgrpc.Channel

	wrapperUnary  service_type.MiddleNext
	wrapperStream service_type.MiddleNext
}

func (t *implService) Debug() fiber2.Router {
	return t.gw.Get().Group("/debug")
}

func (t *implService) plugins() []plugin.Plugin { return t.pluginList }

func (t *implService) middleware(mid service_type.Middleware) {
	t.middlewares = append(t.middlewares, mid)
}
func (t *implService) BeforeStarts(f ...func())    { t.beforeStarts = append(t.beforeStarts, f...) }
func (t *implService) BeforeStops(f ...func())     { t.beforeStops = append(t.beforeStops, f...) }
func (t *implService) AfterStarts(f ...func())     { t.afterStarts = append(t.afterStarts, f...) }
func (t *implService) AfterStops(f ...func())      { t.afterStops = append(t.afterStops, f...) }
func (t *implService) Plugin(plugin plugin.Plugin) { t.pluginList = append(t.pluginList, plugin) }

func (t *implService) init() error {
	defer xerror.RespExit()

	// 依赖对象注入
	inject.Inject(t)

	// 编码注册
	encoding3.Each(func(_ string, cdc encoding3.Codec) {
		encoding.RegisterCodec(cdc)
	})

	// 配置解析
	_ = config.Decode(Name, &t.cfg)

	// 网关初始化
	xerror.Panic(t.gw.Build(t.cfg.Gw))

	// 注册系统middleware
	t.srv.UnaryInterceptor(t.handlerUnaryMiddle(t.middlewares))
	t.srv.StreamInterceptor(t.handlerStreamMiddle(t.middlewares))

	// grpc serve初始化
	xerror.Panic(t.srv.Build(t.cfg.Grpc))

	// 加载inproc的middleware
	t.inproc.WithServerUnaryInterceptor(t.handlerUnaryMiddle(t.middlewares))
	t.inproc.WithServerStreamInterceptor(t.handlerStreamMiddle(t.middlewares))

	t.gw.Get().Use(t.handlerHttpMiddle(t.middlewares))

	// 初始化 handlers
	for _, srv := range t.ServiceDesc() {
		// service handler依赖对象注入
		logutil.LogOrPanic(t.L, "Handler Dependency Injection",
			func() error {
				inject.Inject(srv.Handler)
				return nil
			},
			zap.String("handler", fmt.Sprintf("%#v", srv)),
		)

		if h, ok := srv.Handler.(service_type.Handler); ok {
			logutil.LogOrPanic(t.L, "Handler initCfg", func() error {
				return xerror.Try(func() {
					// register router
					h.Router(t.gw.Get())
					// service handler init
					t.AfterStops(h.Init())
				})
			})
		}
	}
	return nil
}

func (t *implService) Flags(flags ...cli.Flag) {
	if len(flags) == 0 {
		return
	}

	t.cmd.Flags = append(t.cmd.Flags, flags...)
}

func (t *implService) command() *cli.Command { return t.cmd }

func (t *implService) RegisterMatcher(priority int64, matches ...func(io.Reader) bool) func() net.Listener {
	var matchList []cmux.Matcher
	for i := range matches {
		matchList = append(matchList, matches[i])
	}
	return t.net.handler(priority, matchList...)
}

func (t *implService) ServiceDesc() []service_type.Desc {
	return t.services
}

func (t *implService) Options() service_type.Options {
	//TODO implement me
	panic("implement me")
}

func (t *implService) GrpcClientInnerConn() grpc.ClientConnInterface { return t.inproc }

func (t *implService) RegisterService(desc service_type.Desc) {
	xerror.Assert(desc.Handler == nil, "[handler] is nil")

	t.srv.Get().RegisterService(&desc.ServiceDesc, desc.Handler)
	// 进程内grpc serve注册
	t.inproc.RegisterService(&desc.ServiceDesc, desc.Handler)
	t.services = append(t.services, desc)

	if h, ok := desc.Handler.(service_type.Handler); ok {
		t.Flags(h.Flags()...)
	}
}

func (t *implService) serve() error { return t.net.Serve() }

func (t *implService) start() (gErr error) {
	defer xerror.RespErr(&gErr)

	logutil.OkOrPanic(t.L, "service before-start", func() error {
		var beforeList []func()
		for _, p := range plugin.All() {
			beforeList = append(beforeList, p.BeforeStarts()...)
		}
		beforeList = append(beforeList, t.beforeStarts...)
		for i := range beforeList {
			t.L.Sugar().Infof("running %s", stack.Func(beforeList[i]))
			xerror.PanicF(xerror.Try(beforeList[i]), stack.Func(beforeList[i]))
		}
		return nil
	})

	logutil.OkOrPanic(t.L, "service start", func() error {
		t.L.Sugar().Infof("Server Listening on http://%s:%d", netutil.GetLocalIP(), netutil.MustGetPort(runtime.Addr))
		var grpcLn = t.net.Grpc()

		// 启动grpc服务
		syncx.GoDelay(func() {
			t.L.Info("[grpc] Server Starting")
			logutil.LogOrErr(t.L, "[grpc] Server Stop", func() error {
				if err := t.srv.Get().Serve(grpcLn()); err != nil &&
					err != cmux.ErrListenerClosed &&
					!errors.Is(err, http.ErrServerClosed) &&
					!errors.Is(err, net.ErrClosed) {
					return err
				}
				return nil
			})
		})

		var gwLn = t.net.HTTP1()
		// 启动grpc网关
		syncx.GoDelay(func() {
			t.L.Info("[grpc-gw] Server Starting")
			logutil.LogOrErr(t.L, "[grpc-gw] Server Stop", func() error {
				if err := t.gw.Get().Server().Serve(gwLn()); err != nil &&
					!errors.Is(err, cmux.ErrListenerClosed) &&
					!errors.Is(err, http.ErrServerClosed) &&
					!errors.Is(err, net.ErrClosed) {
					return err
				}
				return nil
			})
		})

		// 启动net网络
		syncx.GoDelay(func() {
			t.L.Info("[cmux] Server Starting")
			logutil.LogOrErr(t.L, "[cmux] Server Stop", func() error {
				if err := t.serve(); err != nil &&
					!errors.Is(err, http.ErrServerClosed) &&
					!errors.Is(err, net.ErrClosed) {
					return err
				}
				return nil
			})
		})
		return nil
	})

	logutil.OkOrPanic(t.L, "service after-start", func() error {
		var afterList []func()
		for _, p := range plugin.All() {
			afterList = append(afterList, p.AfterStarts()...)
		}
		afterList = append(afterList, t.afterStarts...)
		for i := range afterList {
			t.L.Sugar().Infof("running %s", stack.Func(afterList[i]))
			xerror.PanicF(xerror.Try(afterList[i]), stack.Func(afterList[i]))
		}
		return nil
	})
	return nil
}

func (t *implService) stop() (err error) {
	defer xerror.RespErr(&err)

	logutil.OkOrErr(t.L, "service before-stop", func() error {
		var beforeList []func()
		for _, p := range plugin.All() {
			beforeList = append(beforeList, p.BeforeStops()...)
		}
		beforeList = append(beforeList, t.beforeStops...)
		for i := range beforeList {
			t.L.Sugar().Infof("running %s", stack.Func(beforeList[i]))
			xerror.PanicF(xerror.Try(beforeList[i]), stack.Func(beforeList[i]))
		}
		return nil
	})

	logutil.LogOrErr(t.L, "[grpc] GracefulStop", func() error {
		t.srv.Get().GracefulStop()
		xerror.Panic(t.gw.Get().Shutdown())
		xerror.Panic(t.net.Close())
		return nil
	})

	logutil.OkOrErr(t.L, "service after-stop", func() error {
		var afterList []func()
		for _, p := range plugin.All() {
			afterList = append(afterList, p.AfterStops()...)
		}
		afterList = append(afterList, t.afterStops...)
		for i := range afterList {
			t.L.Sugar().Infof("running %s", stack.Func(afterList[i]))
			xerror.PanicF(xerror.Try(afterList[i]), stack.Func(afterList[i]))
		}
		return nil
	})

	return
}
