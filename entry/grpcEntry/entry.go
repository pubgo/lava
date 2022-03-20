package grpcEntry

import (
	"context"
	"errors"
	"fmt"
	"github.com/pubgo/lava/service/service_type"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/google/uuid"
	grpcMiddle "github.com/grpc-ecosystem/go-grpc-middleware"
	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pubgo/dix"
	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
	"github.com/soheilhy/cmux"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"github.com/pubgo/lava/config"
	encoding3 "github.com/pubgo/lava/encoding"
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/entry/base"
	"github.com/pubgo/lava/entry/grpcEntry/grpc-gw"
	"github.com/pubgo/lava/entry/grpcEntry/grpcs"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/plugins/registry"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/version"
)

func New(name string) Entry { return newEntry(name) }
func newEntry(name string) *grpcEntry {
	var g = &grpcEntry{
		Entry:  base.New(name),
		srv:    grpcs.New(name),
		gw:     grpc_gw.New(name),
		inproc: &inprocgrpc.Channel{},
		cfg: Cfg{
			name:                 name,
			hostname:             runtime.Hostname,
			id:                   uuid.New().String(),
			Grpc:                 grpcs.GetDefaultCfg(),
			Gw:                   grpc_gw.DefaultCfg(),
			RegisterTTL:          time.Minute,
			RegisterInterval:     time.Second * 30,
			SleepAfterDeRegister: time.Second * 2,
		},
	}

	g.OnInit(func() {
		defer xerror.RespExit()

		// 编码注册
		encoding3.Each(func(_ string, cdc encoding3.Codec) {
			encoding.RegisterCodec(cdc)
		})

		// 配置解析
		_ = config.Decode(Name, &g.cfg)

		// 注册中心加载
		g.registry = registry.Default()

		// 网关初始化
		xerror.Panic(g.gw.Build(g.cfg.Gw))

		// 注册系统middleware
		g.srv.UnaryInterceptor(g.handlerUnaryMiddle(g.Options().Middlewares))
		g.srv.StreamInterceptor(g.handlerStreamMiddle(g.Options().Middlewares))

		// 注册自定义middleware
		g.srv.UnaryInterceptor(g.unaryServerInterceptors...)
		g.srv.StreamInterceptor(g.streamServerInterceptors...)

		// 加载inproc的middleware
		g.inproc.WithServerUnaryInterceptor(grpcMiddle.ChainUnaryServer(
			append([]grpc.UnaryServerInterceptor{
				g.handlerUnaryMiddle(g.Options().Middlewares)}, g.unaryServerInterceptors...)...,
		))
		g.inproc.WithServerStreamInterceptor(grpcMiddle.ChainStreamServer(
			append([]grpc.StreamServerInterceptor{
				g.handlerStreamMiddle(g.Options().Middlewares)}, g.streamServerInterceptors...)...,
		))

		// grpc serve初始化
		xerror.Panic(g.srv.Build(g.cfg.Grpc))

		// 初始化handlers
		for _, srv := range g.Options().Handlers {
			// 注册grpc handler
			logutil.LogOrPanic(logs.L(), "Grpc Handler Register", func() error {
				// 注册handler, 同时注册到grpc和inproc
				return registerGrpc(g, srv)
			})

			// 注册gateway handler
			// 进程内通信, 通过inproc绑定grpc serve和client
			logutil.LogOrPanic(logs.L(), "Gateway Handler Register ", func() error {
				return registerGw(context.Background(), g.gw.Get(), g.inproc, srv)
			})

			// Handler对象注入
			logutil.LogOrPanic(logs.L(), "Handler Dependency Injection",
				func() error {
					err := dix.Inject(srv)
					if err == nil {
						return nil
					}

					// 对象详情
					q.Q(srv)

					// 当前依赖注入对象graph
					fmt.Println(dix.Graph())
					return err
				},
				zap.String("handler", fmt.Sprintf("%#v", srv)),
			)

			// Handler初始化
			logutil.LogOrPanic(logs.L(), "Handler initCfg", func() error { return xerror.Try(srv.Init) })
		}
	})

	return g
}

var _ Entry = (*grpcEntry)(nil)
var logs = logging.Component(Name)

type grpcEntry struct {
	*base.Entry
	cfg         Cfg
	mux         cmux.CMux
	srv         grpcs.Builder
	gw          grpc_gw.Builder
	middlewares []service_type.Middleware

	// inproc Channel is used to serve grpc gateway
	inproc *inprocgrpc.Channel

	registry    registry.Registry
	registered  atomic.Bool
	registryMap map[string][]*registry.Endpoint

	cancelRegister context.CancelFunc

	wrapperUnary  func(ctx context.Context, req service_type.Request, rsp func(response service_type.Response) error) error
	wrapperStream func(ctx context.Context, req service_type.Request, rsp func(response service_type.Response) error) error

	unaryServerInterceptors  []grpc.UnaryServerInterceptor
	streamServerInterceptors []grpc.StreamServerInterceptor
}

func (g *grpcEntry) Mux() *gw.ServeMux              { return g.gw.Get() }
func (g *grpcEntry) Conn() grpc.ClientConnInterface { return g.inproc }

func (g *grpcEntry) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	g.srv.Get().RegisterService(desc, impl)
	// 进程内grpc serve注册
	g.inproc.RegisterService(desc, impl)
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

		logs.WithErr(err).Error("grpcEntry cmux handleError")
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
		advt = runtime.Addr
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
		logs.L().Info("Registering Node", zap.String("id", node.Id), zap.String("name", g.cfg.name))
	}

	// registry options
	opts := []registry.RegOpt{registry.TTL(g.cfg.RegisterTTL)}
	logutil.LogOrPanic(logs.L(), "[grpc] register", func() error { return g.registry.Register(services, opts...) })

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

	logutil.LogOrErr(logs.L(), "deregister node", func() error { return g.registry.Deregister(services) },
		zap.String("id", node.Id),
	)

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
	logutil.LogOrErr(logs.L(), "[grpc] server deRegister", g.deRegister)

	// Add sleep for those requests which have selected this port.
	time.Sleep(g.cfg.SleepAfterDeRegister)

	logutil.LogOrErr(logs.L(), "[grpc] GracefulStop", func() error {
		g.srv.Get().GracefulStop()
		return nil
	})
	return
}

func (g *grpcEntry) Register(handler entry.Handler) {
	defer xerror.RespExit()
	xerror.Assert(handler == nil, "[handler] should not be nil")
	xerror.Assert(!findGrpcHandle(handler).IsValid(), "register [%#v] 没有找到匹配的interface", handler)
	g.RegisterHandler(handler)
}

func (g *grpcEntry) Start() (gErr error) {
	defer xerror.RespErr(&gErr)

	logs.S().Infof("Server Listening on http://%s:%d", netutil.GetLocalIP(), netutil.MustGetPort(runtime.Addr))
	ln := xerror.PanicErr(netutil.Listen(runtime.Addr)).(net.Listener)

	// mux server acts as a reverse-proxy between HTTP and GRPC backends.
	g.mux = cmux.New(ln)
	g.mux.SetReadTimeout(g.cfg.Gw.Timeout)
	g.handleError()

	// 启动grpc服务
	syncx.GoDelay(func() {
		logs.L().Info("[grpc] Server Starting")
		logutil.LogOrErr(logs.L(), "[grpc] Server Stop", func() error {
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
			logutil.LogOrErr(logs.L(), "[grpc-gw] Shutdown", func() error {
				if err := s.Shutdown(context.Background()); err != nil && !errors.Is(err, net.ErrClosed) {
					return err
				}
				return nil
			})
		})

		logs.L().Info("[grpc-gw] Server Starting")
		logutil.LogOrErr(logs.L(), "[grpc-gw] Server Stop", func() error {
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
		logs.L().Info("[cmux] Server Starting")
		logutil.LogOrErr(logs.L(), "[cmux] Server Stop", func() error {
			if err := g.serve(); err != nil &&
				!errors.Is(err, http.ErrServerClosed) &&
				!errors.Is(err, net.ErrClosed) {
				return err
			}
			return nil
		})
	})

	// register self
	logutil.LogOrPanic(logs.L(), "[grpc] start to register", g.register)

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
				logutil.LogOrErr(logs.L(), "service register",
					g.register,
					zap.String("registry", g.registry.String()),
					zap.String("interval", interval.String()),
				)
			case <-ctx.Done():
				logs.L().Info("service register cancelled")
				return
			}
		}
	})

	return nil
}
