package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/service/internal/fiber_builder"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	fiber2 "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pubgo/xerror"
	"github.com/soheilhy/cmux"
	"github.com/urfave/cli/v2"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"github.com/pubgo/lava/config"
	encoding3 "github.com/pubgo/lava/encoding"
	"github.com/pubgo/lava/entry/grpcEntry/grpcs"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/plugins/registry"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/types"
)

var _ Service = (*grpcEntry)(nil)

func newEntry(name string, desc string) *grpcEntry {
	var g = &grpcEntry{
		cmd:    &cli.Command{Name: name, Usage: desc},
		srv:    grpcs.New(name),
		gw:     fiber_builder.New(),
		inproc: &inprocgrpc.Channel{},
		net: &netCfg{
			Addr:        "0.0.0.0",
			Port:        8080,
			ReadTimeout: time.Second * 2,
			HandleError: func(err error) bool {
				zap.L().Named("cmux").Error("HandleError", logutil.ErrField(err)...)
				return false
			},
		},
		cfg: Cfg{
			name:     name,
			hostname: runtime.Hostname,
			id:       uuid.New().String(),
			Grpc:     grpcs.GetDefaultCfg(),
			Gw:       fiber_builder.Cfg{},
		},
	}

	return g
}

var logs = logging.Component(Name)

type grpcEntry struct {
	beforeStarts []func()
	afterStarts  []func()
	beforeStops  []func()
	afterStops   []func()
	pluginList   []plugin.Plugin
	middlewares  []types.Middleware
	services     []Desc

	cmd *cli.Command

	net *netCfg

	cfg Cfg
	mux cmux.CMux
	srv grpcs.Builder
	gw  fiber_builder.Builder

	// inproc Channel is used to serve grpc gateway
	inproc *inprocgrpc.Channel

	registered  atomic.Bool
	registryMap map[string][]*registry.Endpoint

	wrapperUnary  types.MiddleNext
	wrapperStream types.MiddleNext
}

func (t *grpcEntry) Debug() fiber2.Router {
	return t.gw.Get().Group("/debug")
}

func (t *grpcEntry) Admin() fiber2.Router {
	return t.gw.Get().Group("/admin")
}

func (t *grpcEntry) plugins() []plugin.Plugin { return t.pluginList }

func (t *grpcEntry) middleware(mid types.Middleware) { t.middlewares = append(t.middlewares, mid) }
func (t *grpcEntry) BeforeStarts(f ...func())        { t.beforeStarts = append(t.beforeStarts, f...) }
func (t *grpcEntry) BeforeStops(f ...func())         { t.beforeStops = append(t.beforeStops, f...) }
func (t *grpcEntry) AfterStarts(f ...func())         { t.afterStarts = append(t.afterStarts, f...) }
func (t *grpcEntry) AfterStops(f ...func())          { t.afterStops = append(t.afterStops, f...) }
func (t *grpcEntry) Plugin(plugin plugin.Plugin)     { t.pluginList = append(t.pluginList, plugin) }

func (t *grpcEntry) init() error {
	defer xerror.RespExit()

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

	// 加载inproc的middleware
	t.inproc.WithServerUnaryInterceptor(t.handlerUnaryMiddle(t.middlewares))
	t.inproc.WithServerStreamInterceptor(t.handlerStreamMiddle(t.middlewares))

	// grpc serve初始化
	xerror.Panic(t.srv.Build(t.cfg.Grpc))

	// 初始化 handlers
	for _, srv := range t.ServiceDesc() {
		// Handler对象注入
		logutil.LogOrPanic(logs.L(), "Handler Dependency Injection",
			func() error {
				inject.Inject(srv.Handler)
				return nil
			},
			zap.String("handler", fmt.Sprintf("%#v", srv)),
		)

		switch srv.Handler.(type) {
		case Handler:
			var h = srv.Handler.(Handler)
			// Handler初始化
			logutil.LogOrPanic(logs.L(), "Handler initCfg", func() error {
				return xerror.Try(func() {
					t.AfterStops()
					h.Init()
					h.Flags()
				})
			})
		}
	}
	return nil
}

func (t *grpcEntry) Flags(flags ...cli.Flag) {
	if len(flags) == 0 {
		return
	}

	t.cmd.Flags = append(t.cmd.Flags, flags...)
}

func (t *grpcEntry) command() *cli.Command { return t.cmd }

func (t *grpcEntry) Description(usage string, description ...string) {
	t.cmd.Usage = usage

	if len(description) > 0 {
		t.cmd.UsageText = description[1]
	}

	if len(description) > 1 {
		t.cmd.Description = description[2]
	}

	return
}

func (t *grpcEntry) RegisterMatcher(priority int64, matches ...func(io.Reader) bool) func() net.Listener {
	var matchList []cmux.Matcher
	for i := range matches {
		matchList = append(matchList, matches[i])
	}
	return t.net.handler(priority, matchList...)
}

func (t *grpcEntry) ServiceDesc() []Desc {
	return t.services
}

func (t *grpcEntry) Options() Options {
	//TODO implement me
	panic("implement me")
}

func (t *grpcEntry) GrpcClientInnerConn() grpc.ClientConnInterface { return t.inproc }

func (t *grpcEntry) RegisterService(desc Desc) {
	t.srv.Get().RegisterService(&desc.ServiceDesc, desc.Handler)
	// 进程内grpc serve注册
	t.inproc.RegisterService(&desc.ServiceDesc, desc.Handler)
	t.services = append(t.services, desc)

	switch desc.Handler.(type) {
	case Handler:
		var h = desc.Handler.(Handler)
		t.Flags(h.Flags()...)
	}
}

func (t *grpcEntry) serve() error { return t.mux.Serve() }
func (t *grpcEntry) handleError() {
	t.mux.HandleError(func(err error) bool {
		if errors.Is(err, net.ErrClosed) {
			return true
		}

		logs.WithErr(err).Error("grpcEntry cmux handleError")
		return false
	})
}

func (t *grpcEntry) matchAny() net.Listener   { return t.mux.Match(cmux.Any()) }
func (t *grpcEntry) matchHttp1() net.Listener { return t.mux.Match(cmux.HTTP1()) }
func (t *grpcEntry) matchHttp2() net.Listener {
	return t.mux.Match(
		cmux.HTTP2(),
		cmux.HTTP2HeaderFieldPrefix("content-type", "application/grpc"),
		cmux.HTTP2HeaderFieldPrefix("content-type", "application/grpc+proto"),
		cmux.HTTP2HeaderFieldPrefix("content-type", "application/grpc+json"),
	)
}

func (t *grpcEntry) stop() (err error) {
	defer xerror.RespErr(&err)

	logutil.OkOrErr(logs.L(), "before-stop running", func() error {
		var beforeList []func()
		for _, p := range plugin.All() {
			beforeList = append(beforeList, p.BeforeStops()...)
		}
		beforeList = append(beforeList, ent.Options().BeforeStops...)
		for i := range beforeList {
			logs.S().Infof("running %s", stack.Func(beforeList[i]))
			xerror.PanicF(xerror.Try(beforeList[i]), stack.Func(beforeList[i]))
		}
		return nil
	})

	logutil.LogOrErr(logs.L(), "[grpc] GracefulStop", func() error {
		t.srv.Get().GracefulStop()
		return nil
	})

	logutil.OkOrErr(logs.L(), "after-stop running", func() error {
		var afterList []func()
		for _, p := range plugin.All() {
			afterList = append(afterList, p.AfterStops()...)
		}
		afterList = append(afterList, ent.Options().AfterStops...)
		for i := range afterList {
			logs.S().Infof("running %s", stack.Func(afterList[i]))
			xerror.PanicF(xerror.Try(afterList[i]), stack.Func(afterList[i]))
		}
		return nil
	})

	return
}

func (t *grpcEntry) start() (gErr error) {
	defer xerror.RespErr(&gErr)

	logutil.OkOrPanic(logs.L(), "before-start running", func() error {
		var beforeList []func()
		for _, p := range plugin.All() {
			beforeList = append(beforeList, p.BeforeStarts()...)
		}
		beforeList = append(beforeList, ent.Options().BeforeStarts...)
		for i := range beforeList {
			logs.S().Infof("running %s", stack.Func(beforeList[i]))
			xerror.PanicF(xerror.Try(beforeList[i]), stack.Func(beforeList[i]))
		}
		return nil
	})
	logutil.OkOrPanic(logs.L(), "server start", ent.Start)
	logutil.OkOrPanic(logs.L(), "after-start running", func() error {
		var afterList []func()
		for _, p := range plugin.All() {
			afterList = append(afterList, p.AfterStarts()...)
		}
		afterList = append(afterList, ent.Options().AfterStarts...)
		for i := range afterList {
			logs.S().Infof("running %s", stack.Func(afterList[i]))
			xerror.PanicF(xerror.Try(afterList[i]), stack.Func(afterList[i]))
		}
		return nil
	})

	logs.S().Infof("Server Listening on http://%s:%d", netutil.GetLocalIP(), netutil.MustGetPort(runtime.Addr))
	ln := xerror.PanicErr(netutil.Listen(runtime.Addr)).(net.Listener)

	// mux server acts as a reverse-proxy between HTTP and GRPC backends.
	t.mux = cmux.New(ln)
	t.mux.SetReadTimeout(t.cfg.Gw.Timeout)
	t.handleError()

	// 启动grpc服务
	syncx.GoDelay(func() {
		logs.L().Info("[grpc] Server Starting")
		logutil.LogOrErr(logs.L(), "[grpc] Server Stop", func() error {
			if err := t.srv.Get().Serve(t.matchHttp2()); err != nil &&
				err != cmux.ErrListenerClosed &&
				!errors.Is(err, http.ErrServerClosed) &&
				!errors.Is(err, net.ErrClosed) {
				return err
			}
			return nil
		})
	})

	// 启动grpc网关
	syncx.GoDelay(func() {
		var s = http.Server{Handler: t.gw.Get()}
		// grpc服务关闭之前关闭gateway
		t.BeforeStop(func() {
			logutil.LogOrErr(logs.L(), "[grpc-gw] Shutdown", func() error {
				if err := s.Shutdown(context.Background()); err != nil && !errors.Is(err, net.ErrClosed) {
					return err
				}
				return nil
			})
		})

		logs.L().Info("[grpc-gw] Server Starting")
		logutil.LogOrErr(logs.L(), "[grpc-gw] Server Stop", func() error {
			if err := s.Serve(t.matchHttp1()); err != nil &&
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
		logs.L().Info("[cmux] Server Starting")
		logutil.LogOrErr(logs.L(), "[cmux] Server Stop", func() error {
			if err := t.serve(); err != nil &&
				!errors.Is(err, http.ErrServerClosed) &&
				!errors.Is(err, net.ErrClosed) {
				return err
			}
			return nil
		})
	})
	return nil
}
