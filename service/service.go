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

func New(name string, desc string) service_type.Service {
	return newService(name, desc)
}

func newService(name string, desc string) *implService {
	var g = &implService{
		cmd:    &cli.Command{Name: name, Usage: desc},
		srv:    grpc_builder.New(),
		gw:     fiber_builder.New(),
		inproc: &inprocgrpc.Channel{},
		opts: service_type.Options{
			Name: name,
		},
		debug: fiber2.New(),
		net: &netCfg{
			Addr:        "0.0.0.0",
			Port:        8080,
			ch:          make(chan struct{}),
			ReadTimeout: time.Minute * 2,
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

	opts service_type.Options

	cfg   Cfg
	srv   grpc_builder.Builder
	gw    fiber_builder.Builder
	debug *fiber2.App

	// inproc Channel is used to serve grpc gateway
	inproc *inprocgrpc.Channel

	wrapperUnary  service_type.MiddleNext
	wrapperStream service_type.MiddleNext
}

func (t *implService) Middleware(mid service_type.Middleware) {
	if mid == nil {
		return
	}

	t.middlewares = append(t.middlewares, mid)
}

func (t *implService) Middlewares() []service_type.Middleware { return t.middlewares }

func (t *implService) Debug() fiber2.Router { return t.debug }

func (t *implService) plugins() []plugin.Plugin { return t.pluginList }

func (t *implService) middleware(mid service_type.Middleware) {
	if mid == nil {
		return
	}

	t.middlewares = append(t.middlewares, mid)
}

func (t *implService) BeforeStarts(f ...func()) { t.beforeStarts = append(t.beforeStarts, f...) }
func (t *implService) BeforeStops(f ...func())  { t.beforeStops = append(t.beforeStops, f...) }
func (t *implService) AfterStarts(f ...func())  { t.afterStarts = append(t.afterStarts, f...) }
func (t *implService) AfterStops(f ...func())   { t.afterStops = append(t.afterStops, f...) }
func (t *implService) Plugin(plugin plugin.Plugin) {
	if plugin == nil {
		return
	}
	t.pluginList = append(t.pluginList, plugin)
}

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

	t.initDebug()

	// 网关初始化
	xerror.Panic(t.gw.Build(t.cfg.Gw))
	t.gw.Get().Use(func(ctx *fiber2.Ctx) error {
		fmt.Println(ctx.Request().RequestURI())
		return ctx.Next()
	})
	t.gw.Get().Use(t.handlerHttpMiddle(t.middlewares))
	t.gw.Get().Mount("/", t.debug)
	//pretty.Println(t.gw.Get().Stack())

	// 注册系统middleware
	t.srv.UnaryInterceptor(t.handlerUnaryMiddle(t.middlewares))
	t.srv.StreamInterceptor(t.handlerStreamMiddle(t.middlewares))

	// grpc serve初始化
	xerror.Panic(t.srv.Build(t.cfg.Grpc))

	// 加载inproc的middleware
	t.inproc.WithServerUnaryInterceptor(t.handlerUnaryMiddle(t.middlewares))
	t.inproc.WithServerStreamInterceptor(t.handlerStreamMiddle(t.middlewares))

	// 初始化 handlers
	for _, srv := range t.ServiceDesc() {
		// service handler依赖对象注入
		xerror.Panic(xerror.Try(func() {
			inject.Inject(srv.Handler)
			t.L.Info("Handler Dependency Injection", zap.String("handler", fmt.Sprintf("%#v", srv.Handler)))
		}))

		if h, ok := srv.Handler.(service_type.Handler); ok {
			logutil.LogOrPanic(t.L, "Service Handler Init", func() error {
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

func (t *implService) ServiceDesc() []service_type.Desc { return t.services }

func (t *implService) Options() service_type.Options { return t.opts }

func (t *implService) GrpcClientInnerConn() grpc.ClientConnInterface { return t.inproc }

func (t *implService) RegisterService(desc service_type.Desc) {
	xerror.Assert(desc.Handler == nil, "[handler] is nil")

	t.srv.RegisterService(&desc.ServiceDesc, desc.Handler)
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
			if p == nil {
				continue
			}
			beforeList = append(beforeList, p.BeforeStarts()...)
		}
		beforeList = append(beforeList, t.beforeStarts...)
		for i := range beforeList {
			t.L.Sugar().Infof("running %s", stack.Func(beforeList[i]))
			xerror.PanicF(xerror.Try(beforeList[i]), stack.Func(beforeList[i]))
		}
		return nil
	})

	var grpcLn = t.net.Grpc()
	var gwLn = t.net.HTTP1Fast()
	//var app=fiber2.New()
	//t.net.handler(8, func(reader io.Reader) bool {
	//	br := bufio.NewReader(&io.LimitedReader{R: reader, N: 4096})
	//	l, part, err := br.ReadLine()
	//	t.L.Info(string(l))
	//	if err != nil || part {
	//		logutil.LogOrErr(zap.L(), "ReadLine", func() error { return err })
	//		return false
	//	}
	//
	//	t.L.Info(string(l))
	//	return true
	//})

	// 启动grpc网关
	syncx.GoDelay(func() {
		t.L.Info("[grpc-gw] Server Starting")
		logutil.LogOrErr(t.L, "[grpc-gw] Server Stop", func() error {
			if err := t.gw.Get().Listener(gwLn()); err != nil &&
				!errors.Is(err, cmux.ErrListenerClosed) &&
				!errors.Is(err, http.ErrServerClosed) &&
				!errors.Is(err, net.ErrClosed) {
				return err
			}
			return nil
		})
	})

	logutil.OkOrPanic(t.L, "service start", func() error {
		t.L.Sugar().Infof("Server Listening on http://%s:%d", netutil.GetLocalIP(), netutil.MustGetPort(runtime.Addr))

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
			if p == nil {
				continue
			}
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
			if p == nil {
				continue
			}
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
			if p == nil {
				continue
			}
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
