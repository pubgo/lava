package grpcc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/clients/grpcc/resolver"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/logging/logkey"
	"github.com/pubgo/lava/core/logging/logutil"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/runtime"
)

var clients sync.Map
var logs = logging.Component(Name)

func InitClient(srv string, opts ...func(cfg *Cfg)) {
	defer xerror.RespExit()

	var cfg = DefaultCfg(opts...)
	xerror.Panic(cfg.Check())

	if cfg.Group == "" {
		cfg.Group = consts.KeyDefault
	}

	var srvId = fmt.Sprintf("%s.%s", srv, cfg.Group)
	logs.L().Info("grpc client init", zap.String(logkey.Service, srvId))
	if val, ok := clients.LoadOrStore(srvId, NewClient(srv, cfg)); ok && val != nil {
		return
	}

	xerror.Assert(cfg.clientType == nil, "grpc clientType is nil")

	// 依赖注入
	inject.Register(cfg.clientType, func(obj inject.Object, field inject.Field) (interface{}, bool) {
		var conn, ok = clients.Load(fmt.Sprintf("%s.%s", srv, field.Name()))
		if ok {
			return cfg.newClient(conn.(grpc.ClientConnInterface)), true
		}

		logs.L().Error("grpc service not found", zap.String(logkey.Service, srvId))
		return nil, false
	})
}

func New(service string, opts ...func(cfg *Cfg)) *Client {
	return NewClient(service, DefaultCfg(opts...))
}

// NewClient build grpc client
func NewClient(service string, cfg Cfg) *Client {
	var name = service

	// 127.0.0.1,127.0.0.1,127.0.0.1;127.0.0.1
	var host = extractHostFromHostPort(service)
	if strings.Contains(service, ",") || net.ParseIP(host) != nil || host == "localhost" {
		cfg.buildScheme = resolver.DirectScheme
	}

	switch cfg.buildScheme {
	case resolver.DiscovScheme:
		service = resolver.BuildDiscovTarget(service, cfg.registry)
	case resolver.DirectScheme:
		service = resolver.BuildDirectTarget(service)
	default:
		service, name = resolver.Interpret(service)
	}

	return &Client{addr: service, cfg: cfg, name: name}
}

var _ grpc.ClientConnInterface = (*Client)(nil)

type Client struct {
	cfg  Cfg
	addr string
	name string
	mu   sync.Mutex
	conn *grpc.ClientConn
}

func (t *Client) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	var conn, err = t.Get()
	if err != nil {
		return xerror.Wrap(err, method, args)
	}

	return xerror.Wrap(conn.Invoke(ctx, method, args, reply, opts...), method)
}

func (t *Client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	var conn, err = t.Get()
	if err != nil {
		return nil, xerror.Wrap(err, method)
	}

	var c, err1 = conn.NewStream(ctx, desc, method, opts...)
	return c, xerror.Wrap(err1, method)
}

// Get new grpc Client
func (t *Client) Get() (_ grpc.ClientConnInterface, gErr error) {
	defer xerror.Resp(func(err xerror.XErr) {
		gErr = err

		if runtime.IsDev() || runtime.IsTest() {
			logutil.Pretty(t)
			logutil.Pretty(err)
		}
	})

	if t.conn != nil {
		return t.conn, nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// 双检, 避免多次创建
	if t.conn != nil {
		return t.conn, nil
	}

	if val := config.GetMap(Name, t.name); val != nil {
		xerror.Panic(val.Decode(&t.cfg))
	}

	if t.cfg.beforeDial != nil {
		t.cfg.beforeDial()
	}

	// 创建grpc client
	ctx, cancel := context.WithTimeout(context.Background(), t.cfg.DialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, t.addr, append(t.cfg.ToOpts(), t.cfg.DialOptions...)...)
	xerror.PanicF(err, "DialContext error, target:%s\n", t.addr)

	if t.cfg.afterDial != nil {
		t.cfg.afterDial()
	}

	t.conn = conn
	return t.conn, nil
}
