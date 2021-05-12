package rpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	grpcMid "github.com/grpc-ecosystem/go-grpc-middleware"
	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry/base"
	"github.com/pubgo/lug/registry"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/soheilhy/cmux"
	"github.com/spf13/pflag"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

var _ Entry = (*grpcEntry)(nil)

type grpcEntry struct {
	*base.Entry
	mux                      cmux.CMux
	exit                     chan chan error
	mu                       sync.RWMutex
	cfg                      Cfg
	registered               atomic.Bool
	srv                      *grpc.Server
	registryMap              map[string][]*registry.Endpoint
	handlers                 []interface{}
	endpoints                []*registry.Endpoint
	opts                     []grpc.ServerOption
	unaryServerInterceptors  []grpc.UnaryServerInterceptor
	streamServerInterceptors []grpc.StreamServerInterceptor
}

func (g *grpcEntry) Health(fn func() error) error {
	return fn()
}

func (g *grpcEntry) InitOpts(opts ...grpc.ServerOption) { g.opts = append(g.opts, opts...) }

// EnableDebug
// https://github.com/grpc/grpc-experiments/tree/master/gdebug
func (g *grpcEntry) EnableDebug() {
	g.BeforeStart(func() {
		grpc.EnableTracing = true
		reflection.Register(g.srv)
		service.RegisterChannelzServiceToServer(g.srv)
	})
}

func (g *grpcEntry) initGw() (gErr error) {
	gw.DefaultContextTimeout = time.Second * 2
	return
}

func (g *grpcEntry) GRPCListener() net.Listener {
	//HTTP2MatchHeaderFieldPrefixSendSettings
	return g.mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
}

func (g *grpcEntry) HTTPListener() net.Listener {
	return g.mux.Match(cmux.HTTP2(), cmux.HTTP1Fast())
}

func (g *grpcEntry) initGrpc() (gErr error) {
	unaryInterceptorList := unaryInterceptors
	unaryInterceptorList = append(unaryInterceptorList, g.unaryServerInterceptors...)

	streamInterceptorList := streamInterceptors
	streamInterceptorList = append(streamInterceptorList, g.streamServerInterceptors...)

	opts := GetDefaultServerOpts()
	opts = append(opts,
		grpc.UnaryInterceptor(grpcMid.ChainUnaryServer(unaryInterceptorList...)),
		grpc.StreamInterceptor(grpcMid.ChainStreamServer(streamInterceptorList...)))

	// 注册中心校验
	g.cfg.registry = registry.Default()

	g.srv = grpc.NewServer(opts...)

	return
}

func (g *grpcEntry) Init() (gErr error) {
	defer xerror.RespErr(&gErr)

	xerror.Panic(g.Entry.Init())
	_ = config.Decode(Name, &g.cfg)
	xerror.Panic(g.initGrpc())
	xerror.Panic(g.initGw())

	return
}

func (g *grpcEntry) register() (err error) {
	defer xerror.RespErr(&err)

	if g.cfg.registry == nil {
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
		advt = g.cfg.Address
	}

	parts := strings.Split(advt, ":")
	if len(parts) > 1 {
		host = strings.Join(parts[:len(parts)-1], ":")
		port, _ = strconv.Atoi(parts[len(parts)-1])
	} else {
		host = parts[0]
	}

	// register service
	node := &registry.Node{
		Id:      g.Options().Name + "-" + getHostname() + "-" + DefaultId,
		Address: fmt.Sprintf("%s:%d", host, port),
		Port:    port,
	}

	node.Metadata["registry"] = g.cfg.registry.String()
	node.Metadata["transport"] = "grpc"

	services := &registry.Service{
		Name:      g.Options().Name,
		Version:   g.Entry.Options().Version,
		Nodes:     []*registry.Node{node},
		Endpoints: g.endpoints,
	}

	if !g.registered.Load() {
		xlog.Infof("Registering node: %s", node.Id)
	}

	// create registry options
	rOpts := []registry.RegOpt{registry.TTL(g.cfg.RegisterTTL)}
	if err := g.cfg.registry.Register(services, rOpts...); err != nil {
		return xerror.WrapF(err, "[grpc] registry register error")
	}

	// already registered? don't need to register subscribers
	if g.registered.Load() {
		return nil
	}

	g.registered.Store(true)
	return nil
}

func (g *grpcEntry) deregister() (err error) {
	defer xerror.RespErr(&err)

	if g.cfg.registry == nil {
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
		Id:      config.Project + "-" + getHostname() + "-" + DefaultId,
		Address: fmt.Sprintf("%s:%d", host, port),
		Port:    port,
	}

	services := &registry.Service{
		Name:    config.Project,
		Version: g.Entry.Options().Version,
		Nodes:   []*registry.Node{node},
	}

	xlog.Infof("DeRegistering node: %s", node.Id)
	if err := g.cfg.registry.DeRegister(services); err != nil {
		return xerror.WrapF(err, "[grpc] registry deregister error")
	}

	if !g.registered.Load() {
		return nil
	}

	g.registered.Store(false)
	return nil
}

func (g *grpcEntry) Stop() (err error) {
	defer xerror.RespErr(&err)

	ch := make(chan error)
	xlog.Info("[ExitProgress] Stop is called, send error chan. before.")
	g.exit <- ch
	xlog.Info("[ExitProgress] Stop is called, send error chan. end.")
	return <-ch
}

func (g *grpcEntry) UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.unaryServerInterceptors = append(g.unaryServerInterceptors, interceptors...)
}

func (g *grpcEntry) StreamInterceptor(interceptors ...grpc.StreamServerInterceptor) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.streamServerInterceptors = append(g.streamServerInterceptors, interceptors...)
}

func (g *grpcEntry) Register(handler interface{}, opts ...Opt) {
	defer xerror.RespDebug()

	xerror.Assert(handler == nil, "[handler] should not be nil")
	xerror.Panic(checkHandle(handler), "[grpc] grpcEntry.Register error")

	g.mu.Lock()
	defer g.mu.Unlock()
	g.handlers = append(g.handlers, handler)
}

// 开启api网关模式
func (g *grpcEntry) startGw() (err error) {
	if g.cfg.Gw.Addr == "" {
		return nil
	}

	gw.DefaultContextTimeout = time.Second * 2

	// 开启api网关模式
	mux := gw.NewServeMux(
		gw.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
			return metadata.MD(r.URL.Query())
		}),

		gw.WithMarshalerOption(gw.MIMEWildcard, &gw.HTTPBodyMarshaler{
			Marshaler: &gw.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:  true,
					UseEnumNumbers: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
	)

	var server = &http.Server{Addr: g.cfg.Gw.Addr, Handler: mux}

	// 注册网关api
	xerror.Panic(registerGw(g.cfg.Address, mux, grpc.WithBlock(), grpc.WithInsecure()))

	g.AfterStart(func() {
		xerror.Exit(fx.GoDelay(time.Second, func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				xlog.Error("Server [GW] Listen Error", xlog.Any("err", err))
			}

			xlog.Info("Server [GW] Closed OK")
		}))

		xlog.Infof("Server [GW] Listening on http://%s", g.cfg.Gw.Addr)
	})

	g.BeforeStop(func() {
		if err := server.Shutdown(context.Background()); err != nil {
			xlog.Error("Server [GW] Shutdown Error", xlog.Any("err", err))
		}
	})

	return nil
}

func (g *grpcEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	g.endpoints = g.endpoints[:0]
	// 初始化routes
	for i := range g.handlers {
		g.endpoints = append(g.endpoints, newRpcHandler(g.handlers[i])...)
		xerror.PanicF(register(g.srv, g.handlers[i]), "[grpc] register error")
	}

	//if g.cfg.Address == "" {
	//	return xerror.New("[grpc] please set address")
	//}

	ts, err := net.Listen("tcp", g.cfg.Address)
	if err != nil {
		return xerror.WrapF(err, "net Listen error, addr:%s", g.cfg.Address)
	}

	xlog.Infof("Server [grpc] Listening on %s", ts.Addr().String())
	g.mu.Lock()
	g.cfg.Address = ts.Addr().String()
	g.mu.Unlock()

	// announce self to the world
	if err := g.register(); err != nil {
		return xerror.WrapF(err, "[grpc] registry try register error")
	}

	_ = fx.Go(func(ctx context.Context) {
		if err := g.srv.Serve(ts); err != nil {
			xlog.Errorf("[grpc] server stop error: %#v", err)
		}
	})

	t := new(time.Ticker)

	// only process if it exists
	if g.cfg.RegisterInterval > time.Duration(0) {
		// new ticker
		t = time.NewTicker(g.cfg.RegisterInterval)
	}

	// return error chan
	var ch chan error

Loop:
	for {
		select {
		// register self on interval
		case <-t.C:
			if err := g.register(); err != nil {
				xlog.Info("[grpc] server register error", xlog.Any("err", err))
			}
		// wait for exit
		case ch = <-g.exit:
			break Loop
		}
	}

	// deregister self
	if err := g.deregister(); err != nil {
		xlog.Info("[grpc] server deregister error", xlog.Any("err", err))
	}

	// Add sleep for those requests which have selected this port.
	time.Sleep(g.cfg.SleepAfterDeregister)

	// wait for waitgroup
	xlog.Info("[ExitProgress] Start wait-group wait.")

	// stop the grpc server
	xlog.Info("[ExitProgress] Start GracefulStop.")
	g.srv.GracefulStop()

	// close transport
	xlog.Info("[ExitProgress] Close transport.")
	ch <- nil
	xlog.Info("[ExitProgress] All is done.")

	return nil
}

func (g *grpcEntry) initFlags() {
	g.Flags(func(flags *pflag.FlagSet) {
		flags.StringVar(&g.cfg.Gw.Addr, "gw_addr", g.cfg.Gw.Addr, "set gateway addr and enable gateway")
	})
}

func newEntry(name string) *grpcEntry {
	ent := &grpcEntry{Entry: base.New(name)}
	ent.initFlags()

	// 服务启动后, 启动网关
	ent.AfterStart(func() { xerror.Panic(ent.startGw()) })
	return ent
}

func New(name string) Entry { return newEntry(name) }
