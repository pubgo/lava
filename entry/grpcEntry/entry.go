package grpcEntry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pubgo/dix"
	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
	"github.com/soheilhy/cmux"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	grpcGw "github.com/pubgo/lava/builder/grpc-gw"
	"github.com/pubgo/lava/builder/grpcs"
	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/entry/base"
	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/logz"
	encoding2 "github.com/pubgo/lava/pkg/encoding"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/plugins/registry"
	"github.com/pubgo/lava/plugins/syncx"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/version"
)

func New(name string) Entry { return newEntry(name) }
func newEntry(name string) *grpcEntry {
	var g = &grpcEntry{
		Entry: base.New(name),
		srv:   grpcs.New(name),
		gw:    grpcGw.New(name),
		cfg: Cfg{
			name:                 name,
			hostname:             env.Hostname,
			id:                   uuid.New().String(),
			Grpc:                 grpcs.GetDefaultCfg(),
			Gw:                   grpcGw.DefaultCfg(),
			RegisterTTL:          time.Minute,
			RegisterInterval:     time.Second * 30,
			SleepAfterDeRegister: time.Second * 2,
		},
	}

	g.OnInit(func() {
		defer xerror.RespExit()

		// encoding register
		encoding2.Each(func(_ string, cdc encoding2.Codec) {
			encoding.RegisterCodec(cdc)
		})

		// grpc_entry配置解析
		_ = config.Decode(Name, &g.cfg)

		// 注册中心初始化
		g.registry = registry.Default()

		// 网关初始化
		xerror.Panic(g.gw.Build(g.cfg.Gw))

		// 默认middleware注册
		g.srv.UnaryInterceptor(g.handlerUnaryMiddle(g.Options().Middlewares))
		g.srv.StreamInterceptor(g.handlerStreamMiddle(g.Options().Middlewares))

		// 自定义middleware注册
		g.srv.UnaryInterceptor(g.unaryServerInterceptors...)
		g.srv.StreamInterceptor(g.streamServerInterceptors...)

		// grpc serve初始化
		xerror.Panic(g.srv.Build(g.cfg.Grpc))

		// 初始化handlers
		for _, srv := range g.Options().Handlers {
			// GRPC注册
			logs.LogAndThrow("Handler Register", func() error {
				return registerGrpc(g.srv.Get(), srv)
			})

			// Handler依赖注入
			logs.LogAndThrow("Handler Dependency Injection",
				func() error {
					if err := dix.Inject(srv); err != nil {
						q.Q(srv)
						fmt.Println(dix.Graph())
						return err
					}
					return nil
				},
				zap.String("handler", fmt.Sprintf("%#v", srv)),
			)

			// Handler初始化
			logs.LogAndThrow("Handler Init", func() error { return xerror.Try(srv.Init) })
		}
	})

	return g
}

var _ Entry = (*grpcEntry)(nil)
var logs = logz.Component(Name)

type grpcEntry struct {
	*base.Entry
	cfg Cfg
	mux cmux.CMux
	srv grpcs.Builder
	gw  grpcGw.Builder

	registry    registry.Registry
	registered  atomic.Bool
	registryMap map[string][]*registry.Endpoint

	cancelRegister context.CancelFunc

	wrapperUnary  func(ctx context.Context, req types.Request, rsp func(response types.Response) error) error
	wrapperStream func(ctx context.Context, req types.Request, rsp func(response types.Response) error) error

	unaryServerInterceptors  []grpc.UnaryServerInterceptor
	streamServerInterceptors []grpc.StreamServerInterceptor
}

func (g *grpcEntry) UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) {
	g.unaryServerInterceptors = append(g.unaryServerInterceptors, interceptors...)
}

func (g *grpcEntry) StreamInterceptor(interceptors ...grpc.StreamServerInterceptor) {
	g.streamServerInterceptors = append(g.streamServerInterceptors, interceptors...)
}

func (g *grpcEntry) serve() error { return g.mux.Serve() }
func (g *grpcEntry) handleError() {
	g.mux.HandleError(func(err error) bool {
		if errors.Is(err, net.ErrClosed) {
			return true
		}

		logs.WithErr(err).Error("grpcEntry mux handleError")
		return false
	})
}

func (g *grpcEntry) matchAny() net.Listener   { return g.mux.Match(cmux.Any()) }
func (g *grpcEntry) matchHttp1() net.Listener { return g.mux.Match(cmux.HTTP1()) }
func (g *grpcEntry) matchHttp2() net.Listener {
	return g.mux.Match(
		cmux.HTTP2(),
		cmux.HTTP2HeaderFieldPrefix("content-type", "application/grpc"),
		cmux.HTTP2HeaderFieldPrefix("content-type", "application/grpc+proto"),
		cmux.HTTP2HeaderFieldPrefix("content-type", "application/grpc+json"),
	)
}

func (g *grpcEntry) register() (err error) {
	defer xerror.RespErr(&err)

	if g.registry == nil {
		return nil
	}

	// parse address for host, port
	var advt, host string
	var port int

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	if len(g.cfg.Advertise) > 0 {
		advt = g.cfg.Advertise
	} else {
		advt = runenv.Addr
	}

	parts := strings.Split(advt, ":")
	if len(parts) > 1 {
		host = strings.Join(parts[:len(parts)-1], ":")
		port, _ = strconv.Atoi(parts[len(parts)-1])
	} else {
		host = parts[0]
	}

	if host == "" {
		host = netutil.GetLocalIP()
	}

	// register service
	node := &registry.Node{
		Port:     port,
		Address:  fmt.Sprintf("%s:%d", host, port),
		Id:       g.cfg.name + "-" + g.cfg.hostname + "-" + g.cfg.id,
		Metadata: make(map[string]string),
	}

	node.Metadata["registry"] = g.registry.String()
	node.Metadata["transport"] = "grpc"

	services := &registry.Service{
		Name:    g.cfg.name,
		Version: version.Version,
		Nodes:   []*registry.Node{node},
	}

	if !g.registered.Load() {
		logs.Infow("Registering Node", logger.Id(node.Id), logger.Name(g.cfg.name))
	}

	// registry options
	opts := []registry.RegOpt{registry.TTL(g.cfg.RegisterTTL)}
	logs.LogAndThrow("[grpc] register", func() error { return g.registry.Register(services, opts...) })

	// already registered? don't need to register subscribers
	if g.registered.Load() {
		return nil
	}

	g.registered.Store(true)
	return nil
}

func (g *grpcEntry) deRegister() (err error) {
	defer xerror.RespErr(&err)

	if g.registry == nil {
		return nil
	}

	var advt, host string
	var port int

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	if len(g.cfg.Advertise) > 0 {
		advt = g.cfg.Advertise
	} else {
		advt = g.cfg.Address
	}

	parts := strings.Split(advt, ":")
	if len(parts) > 1 {
		host = strings.Join(parts[:len(parts)-1], ":")
		port, _ = strconv.Atoi(parts[len(parts)-1])
	} else {
		host = parts[0]
	}

	node := &registry.Node{
		Id:      g.cfg.name + "-" + g.cfg.hostname + "-" + g.cfg.id,
		Address: fmt.Sprintf("%s:%d", host, port),
		Port:    port,
	}

	services := &registry.Service{
		Name:    g.cfg.name,
		Version: version.Version,
		Nodes:   []*registry.Node{node},
	}

	logs.Logs("deregister node",
		func() error { return g.registry.Deregister(services) },
		zap.String("id", node.Id))

	if !g.registered.Load() {
		return nil
	}

	g.registered.Store(false)
	return nil
}

func (g *grpcEntry) Stop() (err error) {
	defer xerror.RespErr(&err)

	if g.cancelRegister != nil {
		g.cancelRegister()
	}

	// deRegister self
	logs.Logs("[grpc] server deRegister", g.deRegister)

	// Add sleep for those requests which have selected this port.
	time.Sleep(g.cfg.SleepAfterDeRegister)

	logs.Logs("[grpc] GracefulStop", func() error {
		g.srv.Get().GracefulStop()
		return nil
	})
	return
}

func (g *grpcEntry) Register(handler entry.InitHandler) {
	defer xerror.RespExit()
	xerror.Assert(handler == nil, "[handler] should not be nil")
	xerror.Assert(!findGrpcHandle(handler).IsValid(), "register [%#v] 没有找到匹配的interface", handler)
	g.RegisterHandler(handler)
}

func (g *grpcEntry) Start() (gErr error) {
	defer xerror.RespErr(&gErr)

	logs.Infof("Server Listening on http://%s:%d", netutil.GetLocalIP(), netutil.MustGetPort(runenv.Addr))
	ln := xerror.PanicErr(netutil.Listen(runenv.Addr)).(net.Listener)

	// mux server acts as a reverse-proxy between HTTP and GRPC backends.
	g.mux = cmux.New(ln)
	g.mux.SetReadTimeout(g.cfg.Gw.Timeout)
	g.handleError()

	// 启动grpc服务
	syncx.GoDelay(func() {
		logs.Info("[grpc] Server Starting")
		logs.Logs("[grpc] Server Stop", func() error {
			if err := g.srv.Get().Serve(g.matchHttp2()); err != nil &&
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
		var s = http.Server{Handler: g.gw.Get()}
		// grpc服务关闭之前关闭gateway
		g.BeforeStop(func() {
			logs.Logs("[grpc-gw] Shutdown", func() error {
				if err := s.Shutdown(context.Background()); err != nil && !errors.Is(err, net.ErrClosed) {
					return err
				}
				return nil
			})
		})

		logs.Info("[grpc-gw] Server Starting")
		logs.Logs("[grpc-gw] Server Stop", func() error {
			if err := s.Serve(g.matchHttp1()); err != nil &&
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
		logs.Info("[cmux] Server Starting")
		logs.Logs("[cmux] Server Stop", func() error {
			if err := g.serve(); err != nil &&
				!errors.Is(err, http.ErrServerClosed) &&
				!errors.Is(err, net.ErrClosed) {
				return err
			}
			return nil
		})
	})

	// 启动本地grpc客户端
	logs.Info("[grpc] Client Connecting")
	conn, err := grpcc.NewDirect(runenv.Addr)
	xerror.Panic(err)
	xerror.Panic(grpcc.HealthCheck(g.cfg.name, conn))
	for _, h := range g.Options().Handlers {
		logs.LogAndThrow("grpc gateway register handler", func() error {
			return registerGw(context.Background(), g.gw.Get(), conn, h)
		})
	}

	// register self
	logs.LogAndThrow("[grpc] try to register self", g.register)

	g.cancelRegister = syncx.GoCtx(func(ctx context.Context) {
		if g.registry == nil {
			return
		}

		var interval = DefaultRegisterInterval

		// only process if it exists
		if g.cfg.RegisterInterval > time.Duration(0) {
			interval = g.cfg.RegisterInterval
		}

		var tick = time.NewTicker(interval)
		defer tick.Stop()

		for {
			select {
			case <-tick.C:
				logs.Logs("service register",
					g.register,
					zap.String("registry", g.registry.String()),
					zap.String("interval", interval.String()),
				)
			case <-ctx.Done():
				logs.Info("service register cancelled")
				return
			}
		}
	})

	return nil
}
