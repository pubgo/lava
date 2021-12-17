package grpcc

import (
	"context"
	"sync"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/ctxutil"
	"github.com/pubgo/lava/resource"
)

func NewDirect(addr string, opts ...func(cfg *Cfg)) (*grpc.ClientConn, error) {
	return getCfg(consts.KeyDefault, opts...).BuildDirect(addr)
}

var clients sync.Map
var mu sync.Mutex

func GetClient(service string, opts ...func(cfg *Cfg)) *Client {
	var fn = func(cfg *Cfg) {}
	if len(opts) > 0 {
		fn = opts[0]
	}

	var cli, ok = clients.Load(service)
	if ok {
		return cli.(*Client)
	}

	mu.Lock()
	defer mu.Unlock()

	// 双检
	cli, ok = clients.Load(service)
	if ok {
		return cli.(*Client)
	}

	cli = &Client{service: service, optFn: fn}
	clients.Store(service, cli)
	return cli.(*Client)
}

var _ resource.Resource = (*Client)(nil)
var _ grpc.ClientConnInterface = (*Client)(nil)

type Client struct {
	cfg     *Cfg
	service string
	mu      sync.Mutex
	optFn   func(cfg *Cfg)
	conn    *grpc.ClientConn
}

func (t *Client) UpdateResObj(val interface{}) { t.conn = val.(*Client).conn }
func (t *Client) Kind() string                 { return Name }
func (t *Client) Close() error                 { return t.conn.Close() }

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

func (t *Client) CheckHealth(opts ...grpc.CallOption) (*grpc_health_v1.HealthCheckResponse, error) {
	ctx := ctxutil.Timeout()
	return grpc_health_v1.NewHealthClient(t).Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: t.service}, opts...)
}

// Get new grpc Client
func (t *Client) Get() (_ grpc.ClientConnInterface, err error) {
	defer xerror.RespErr(&err)

	if t.conn != nil {
		return t.conn, nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// 双检, 避免多次创建
	if t.conn != nil {
		return t.conn, nil
	}

	// 获取服务的自定义配置
	t.cfg = getCfg(t.service)

	// 使用方自定义配置参数
	if t.optFn != nil {
		t.optFn(t.cfg)
	}

	// 创建grpc client
	conn, err := t.cfg.Build(t.service)
	xerror.PanicF(err, "dial %s error", t.service)
	t.conn = conn
	return t.conn, nil
}
