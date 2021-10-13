package grpcEntry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/entry/base"
	"github.com/pubgo/lava/logger"
	grpcGw "github.com/pubgo/lava/pkg/builder/grpc-gw"
	"github.com/pubgo/lava/pkg/builder/grpcs"
	"github.com/pubgo/lava/pkg/ctxutil"
	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/plugins/grpcc"
	"github.com/pubgo/lava/plugins/registry"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/version"

	"github.com/google/uuid"
	"github.com/pubgo/dix"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/soheilhy/cmux"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var _ Entry = (*grpcEntry)(nil)

type grpcEntry struct {
	*base.Entry
	cfg Cfg
	mux cmux.CMux
	srv grpcs.Builder
	gw  grpcGw.Builder

	registry    registry.Registry
	registered  atomic.Bool
	registryMap map[string][]*registry.Endpoint

	handlers  []interface{}
	endpoints []*registry.Endpoint
	client    *grpc.ClientConn

	middleOnce     sync.Once
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

		logs.Error("grpcEntry mux handleError", logger.Err(err), logger.Name(g.cfg.name))
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
		var ip, err = netutil.LocalIP()
		xerror.Panic(err)
		host = ip
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
		Name:      g.cfg.name,
		Version:   version.Version,
		Nodes:     []*registry.Node{node},
		Endpoints: g.endpoints,
	}

	if !g.registered.Load() {
		logs.Info("Registering node", logger.Id(node.Id), logger.Name(g.cfg.name))
	}

	// registry options
	opts := []registry.RegOpt{registry.TTL(g.cfg.RegisterTTL)}
	xerror.Panic(g.registry.Register(services, opts...), "[grpc] register error")

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

	xerror.Panic(g.registry.Deregister(services), "deregister node error")
	logs.Info("deregister node ok", zap.String("id", node.Id))

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
	if err := g.deRegister(); err != nil {
		logs.Info("[grpc] server deRegister error", logger.Err(err))
	}

	// Add sleep for those requests which have selected this port.
	time.Sleep(g.cfg.SleepAfterDeRegister)

	logs.Info("[ExitProgress] Start Shutdown.")
	if err := g.gw.Get().Shutdown(ctxutil.Default().Context()); err != nil && !strings.Contains(err.Error(), net.ErrClosed.Error()) {
		logs.Error("[grpc-gw] Shutdown Error", logger.Err(err))
	} else {
		logs.Info("[ExitProgress] Shutdown Ok.")
	}

	// stop the grpc server
	logs.Info("[ExitProgress] Start GracefulStop.")
	g.srv.Get().GracefulStop()
	logs.Info("[ExitProgress] GracefulStop Ok.")

	return
}

func (g *grpcEntry) Register(handler interface{}, opts ...Opt) {
	defer xerror.RespExit()

	xerror.Assert(handler == nil, "[handler] should not be nil")
	xerror.Assert(!grpcs.FindHandle(handler).IsValid(), "register [%#v] 没有找到匹配的interface", handler)

	g.handlers = append(g.handlers, handler)
	g.endpoints = append(g.endpoints, newRpcHandler(handler)...)
}

func (g *grpcEntry) Start() (gErr error) {
	defer xerror.RespErr(&gErr)

	logs.Sugar().Infof("Server [%s] Listening on http://localhost:%s", g.cfg.name, lavax.GetPort(runenv.Addr))
	ln := xerror.PanicErr(netutil.Listen(runenv.Addr)).(net.Listener)

	// mux server acts as a reverse-proxy between HTTP and GRPC backends.
	g.mux = cmux.New(ln)
	g.handleError()

	// 启动grpc服务
	fx.GoDelay(func() {
		logs.Info("[grpc] Server Starting")
		if err := g.srv.Get().Serve(g.matchHttp2()); err != nil && err != cmux.ErrListenerClosed {
			logs.Error("[grpc] Server Stop", logger.Err(err))
		}
	})

	// 启动grpc网关
	fx.GoDelay(func() {
		logs.Info("[grpc-gw] Server Staring")
		if err := g.gw.Get().Serve(g.matchHttp1()); err != nil && err != cmux.ErrListenerClosed && errors.Is(err, net.ErrClosed) {
			logs.Error("[grpc-gw] Server Stop", logger.Err(err))
		}
	})

	// 启动net网络
	fx.GoDelay(func() {
		logs.Info("[cmux] Server Staring")
		if err := g.serve(); err != nil && !strings.Contains(err.Error(), net.ErrClosed.Error()) {
			logs.Error("[cmux] Server Stop", logger.Err(err))
		}
	})

	// 启动本地grpc客户端
	logs.Info("[grpc] Client Connecting")
	conn, err := grpcc.NewDirect(runenv.Addr)
	xerror.Panic(err)
	g.client = conn
	xerror.Panic(grpcc.HealthCheck(g.cfg.name, g.client))
	for i := range g.handlers {
		xerror.PanicF(g.gw.Register(g.client, g.handlers[i]), "gw register handler error")
	}

	// register self
	xerror.Panic(g.register(), "[grpc] try to register self")

	g.cancelRegister = fx.Go(func(ctx context.Context) {
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
				logs.Sugar().Infow("[grpc] server register", "registry", g.registry.String(), "interval", interval)
				if err := g.register(); err != nil {
					logs.Error("[grpc] server register on interval", logger.Err(err))
				}
			case <-ctx.Done():
				logs.Info("[grpc] register cancelled")
				return
			}
		}
	})

	return nil
}

func newEntry(name string) *grpcEntry {
	var g = &grpcEntry{
		Entry: base.New(name),
		srv:   grpcs.New(name),
		gw:    grpcGw.New(name),
		cfg: Cfg{
			name:                 name,
			hostname:             getHostname(),
			id:                   uuid.New().String(),
			Grpc:                 grpcs.GetDefaultCfg(),
			Gw:                   grpcGw.GetDefaultCfg(),
			RegisterTTL:          time.Minute,
			RegisterInterval:     time.Second * 30,
			SleepAfterDeRegister: time.Second * 2,
		},
	}

	g.OnInit(func() {
		defer xerror.RespExit()

		// encoding register
		grpcs.InitEncoding()

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
		for i := range g.handlers {
			xerror.Panic(dix.Inject(g.handlers[i]))
			xerror.PanicF(grpcs.Register(g.srv.Get(), g.handlers[i]), "grpc register handler error")
		}
	})

	return g
}

func New(name string) Entry { return newEntry(name) }
