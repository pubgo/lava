package golug_grpc

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pubgo/golug/golug_app"
	"github.com/pubgo/golug/golug_entry/golug_base"
	registry "github.com/pubgo/golug/golug_registry"
	"github.com/pubgo/golug/pkg/golug_utils/addr"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess"
	"github.com/spf13/pflag"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"
)

var _ Entry = (*grpcEntry)(nil)

type grpcEntry struct {
	*golug_base.Entry
	exit                     chan chan error
	name                     string
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

func (g *grpcEntry) InitOpts(opts ...grpc.ServerOption) { g.opts = append(g.opts, opts...) }

// EnableDebug
// https://github.com/grpc/grpc-experiments/tree/master/gdebug
func (g *grpcEntry) EnableDebug() {
	grpc.EnableTracing = true
	reflection.Register(g.srv)
	service.RegisterChannelzServiceToServer(g.srv)
}

func (g *grpcEntry) GetDefaultServerOpts() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(g.cfg.MaxReceiveMessageSize),
		grpc.MaxSendMsgSize(g.cfg.MaxSendMessageSize),
	}
}

func (g *grpcEntry) Init() (gErr error) {
	defer xerror.RespErr(&gErr)

	xerror.Panic(g.Entry.Init())

	unaryInterceptorList := unaryInterceptors
	unaryInterceptorList = append(unaryInterceptorList, g.unaryServerInterceptors...)

	streamInterceptorList := streamInterceptors
	streamInterceptorList = append(streamInterceptorList, g.streamServerInterceptors...)

	opts := GetDefaultServerOpts()
	opts = append(opts, g.GetDefaultServerOpts()...)
	opts = append(opts,
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptorList...)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(streamInterceptorList...)))

	// 注册中心校验
	g.cfg.registry = registry.Default

	g.srv = grpc.NewServer(opts...)

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

	addr1, err := addr.Extract(host)
	if err != nil {
		return xerror.WrapF(err, "addr Extract error, host:%s", host)
	}

	// register service
	node := &registry.Node{
		Id:      golug_app.Project + "-" + getHostname() + "-" + DefaultId,
		Address: addr1,
		Port:    port,
	}

	node.Metadata["registry"] = g.cfg.registry.String()
	node.Metadata["transport"] = "grpc"

	services := &registry.Service{
		Name:      golug_app.Project,
		Version:   g.Entry.Options().Version,
		Nodes:     []*registry.Node{node},
		Endpoints: g.endpoints,
	}

	if !g.registered.Load() {
		xlog.Infof("Registering node: %s", node.Id)
	}

	// create registry options
	rOpts := []registry.RegisterOption{registry.TTL(g.cfg.RegisterTTL)}
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

	addr1, err := addr.Extract(host)
	if err != nil {
		return xerror.WrapF(err, "addr Extract error, host:%s", host)
	}

	node := &registry.Node{
		Id:      golug_app.Project + "-" + getHostname() + "-" + DefaultId,
		Address: addr1,
		Port:    port,
	}

	services := &registry.Service{
		Name:    golug_app.Project,
		Version: g.Entry.Options().Version,
		Nodes:   []*registry.Node{node},
	}

	xlog.Infof("DeRegistering node: %s", node.Id)
	if err := g.cfg.registry.Deregister(services); err != nil {
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

func (g *grpcEntry) RegisterUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.unaryServerInterceptors = append(g.unaryServerInterceptors, interceptors...)
}

func (g *grpcEntry) RegisterStreamInterceptor(interceptors ...grpc.StreamServerInterceptor) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.streamServerInterceptors = append(g.streamServerInterceptors, interceptors...)
}

func (g *grpcEntry) Register(handler interface{}, opts ...Option) {
	xerror.Assert(handler == nil, "[handler] should not be nil")
	xerror.ExitF(checkHandle(handler), "[grpc] grpcEntry.Register error")

	g.mu.Lock()
	defer g.mu.Unlock()
	g.handlers = append(g.handlers, handler)
}

// 开启api网关模式
func (g *grpcEntry) startGw() (err error) {
	if g.cfg.GwAddr == "" {
		return nil
	}

	app := fiber.New()

	// 开启api网关模式
	if err := registerGw(
		fmt.Sprintf("localhost:%d", g.Entry.Options().Port),
		app.Group("/")); err != nil {
		return err
	}

	var data []map[string]string
	for i, stacks := range app.Stack() {
		data = append(data, make(map[string]string))
		for _, stack := range stacks {
			if stack == nil {
				continue
			}

			if stack.Path == "/" {
				continue
			}
			data[i][stack.Method] = stack.Path
		}
	}
	fmt.Printf("%#v\n", data)

	return app.Listen(g.cfg.GwAddr)
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

	xprocess.Go(func(ctx context.Context) {
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
		flags.StringVar(&g.cfg.GwAddr, "gw", g.cfg.GwAddr, "set addr and enable gateway mode")
	})
}

func newEntry(name string) *grpcEntry {
	ent := &grpcEntry{Entry: golug_base.New(name)}
	ent.initFlags()

	// 服务启动后, 启动网关
	ent.AfterStart(func() { xerror.Panic(ent.startGw()) })
	ent.OnCfgWithName(Name, &ent.cfg)
	return ent
}

func New(name string) Entry { return newEntry(name) }
