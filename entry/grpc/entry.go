package grpc

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pubgo/lug/app"
	grpcGw "github.com/pubgo/lug/builder/grpc-gw"
	grpcWeb "github.com/pubgo/lug/builder/grpc-web"
	"github.com/pubgo/lug/builder/grpcs"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry/base"
	"github.com/pubgo/lug/pkg/ctxutil"
	"github.com/pubgo/lug/pkg/netutil"
	"github.com/pubgo/lug/registry"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_grpc"
	"github.com/soheilhy/cmux"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
)

var _ Entry = (*grpcEntry)(nil)

type grpcEntry struct {
	*base.Entry
	cfg         Cfg
	web         grpcWeb.Builder
	gw          grpcGw.Builder
	srv         grpcs.Builder
	mux         cmux.CMux
	registry    registry.Registry
	registered  atomic.Bool
	registryMap map[string][]*registry.Endpoint
	handlers    []interface{}
	endpoints   []*registry.Endpoint
}

func (g *grpcEntry) Init(opts ...grpc.ServerOption) { g.srv.Init(opts...) }
func (g *grpcEntry) UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) {
	g.srv.UnaryInterceptor(interceptors...)
}

func (g *grpcEntry) StreamInterceptor(interceptors ...grpc.StreamServerInterceptor) {
	g.srv.StreamInterceptor(interceptors...)
}
func (g *grpcEntry) serve() error             { return g.mux.Serve() }
func (g *grpcEntry) matchAny() net.Listener   { return g.mux.Match(cmux.Any()) }
func (g *grpcEntry) matchHttp1() net.Listener { return g.mux.Match(cmux.HTTP1()) }
func (g *grpcEntry) matchHttp2() net.Listener {
	return g.mux.Match(
		cmux.HTTP2(),
		cmux.HTTP2HeaderFieldPrefix("content-type", "application/grpc"),
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
		Id:      g.cfg.name + "-" + g.cfg.hostname + "-" + g.cfg.id,
		Address: fmt.Sprintf("%s:%d", host, port),
		Port:    port,
	}

	node.Metadata["registry"] = g.registry.String()
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
	if err := g.registry.Register(services, rOpts...); err != nil {
		return xerror.WrapF(err, "[grpc] registry register error")
	}

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
		Name:    app.Project,
		Version: g.Entry.Options().Version,
		Nodes:   []*registry.Node{node},
	}

	xlog.Infof("DeRegistering node: %s", node.Id)
	if err := g.registry.DeRegister(services); err != nil {
		return xerror.WrapF(err, "[grpc] registry deRegister error")
	}

	if !g.registered.Load() {
		return nil
	}

	g.registered.Store(false)
	return nil
}

func (g *grpcEntry) Stop() (err error) {
	defer xerror.RespErr(&err)

	// deRegister self
	if err := g.deRegister(); err != nil {
		xlog.Info("[grpc] server deRegister error", xlog.Any("err", err))
	}

	// Add sleep for those requests which have selected this port.
	time.Sleep(g.cfg.SleepAfterDeregister)

	// stop the grpc server
	xlog.Info("[ExitProgress] Start GracefulStop.")
	g.srv.Get().GracefulStop()
	xlog.Info("[ExitProgress] GracefulStop Ok.")
	xlog.Info("[ExitProgress] Start Shutdown.")
	if err := g.web.Get().Shutdown(ctxutil.Default()); err != nil && !strings.Contains(err.Error(), net.ErrClosed.Error()) {
		xerror.Panic(err)
	}
	xlog.Info("[ExitProgress] Shutdown Ok.")
	return
}

func (g *grpcEntry) initHandler() {
	xerror.RespExit()

	// 初始化routes
	for i := range g.handlers {
		xerror.PanicF(register(g.srv.Get(), g.handlers[i]), "[grpc] register error")
	}
}

func (g *grpcEntry) Register(handler interface{}, opts ...Opt) {
	defer xerror.RespExit()

	xerror.Assert(handler == nil, "[handler] should not be nil")
	xerror.Assert(!checkHandle(handler).IsValid(), "[grpc] grpcEntry.Register error")

	g.handlers = append(g.handlers, handler)
	g.endpoints = append(g.endpoints, newRpcHandler(handler)...)
}

func (g *grpcEntry) Start() (gErr error) {
	defer xerror.RespErr(&gErr)

	xlog.Infof("[%s] Server Listening on %s", app.Project, app.Addr)
	ln, err := netutil.Listen(app.Addr)
	xerror.Panic(err)
	g.mux = cmux.New(ln)

	_ = fx.GoDelay(time.Millisecond*10, func() {
		xlog.Info("Server [grpc] Listening")
		if err := g.srv.Get().Serve(g.matchHttp2()); err != nil && err != cmux.ErrListenerClosed {
			xlog.Error("Server [grpc] Stop error", xlog.Any("err", err))
		}
	})

	_ = fx.GoDelay(time.Millisecond*10, func() {
		xlog.Info("Server [grpc-web] Listening")
		if err := g.web.Get().Serve(g.matchHttp1()); err != nil && err != cmux.ErrListenerClosed {
			xlog.Error("Server [grpc-web] stop error", xlog.Any("err", err))
		}
	})

	xerror.Panic(fx.GoDelay(time.Millisecond*10, func() {
		if err := g.serve(); err != nil && !strings.Contains(err.Error(), net.ErrClosed.Error()) {
			xlog.Error("Server [mux] stop error", xlog.Any("err", err))
		}
	}))

	// announce self to the world
	if err := g.register(); err != nil {
		return xerror.WrapF(err, "[grpc] registry try register error")
	}

	_ = fx.Go(func(ctx context.Context) {
		t := new(time.Ticker)

		// only process if it exists
		if g.cfg.RegisterInterval > time.Duration(0) {
			// new ticker
			t = time.NewTicker(g.cfg.RegisterInterval)
		}

		// register self on interval
		for range t.C {
			if err := g.register(); err != nil {
				xlog.Info("[grpc] server register error", xlog.Any("err", err))
			}
		}
	})

	return nil
}

func newEntry(name string) *grpcEntry {
	var g = &grpcEntry{
		cfg: Cfg{
			RegisterTTL:          time.Minute,
			RegisterInterval:     time.Second * 30,
			SleepAfterDeregister: time.Second * 2,
			hostname:             getHostname(),
			id:                   uuid.New().String(),
			name:                 name,
		},
		Entry: base.New(name),
		srv:   grpcs.New(),
		web:   grpcWeb.New(),
		gw:    grpcGw.New(),
	}

	g.OnInit(func() {
		xlog_grpc.Init(xlog.Named(Name))
		grpcs.InitEncoding()
		_ = config.Decode(Name, &g.cfg)

		// 注册中心校验
		g.registry = registry.Default()

		xerror.Panic(g.srv.Build(g.cfg.Srv))
		xerror.Panic(g.web.Build(g.cfg.Web, g.srv.Get()))
		//xerror.Panic(g.gw.Build(g.cfg.Gw))

		g.initHandler()
	})

	return g
}

func New(name string) Entry { return newEntry(name) }
