package grpcc

import (
	"context"
	"sync"
	"time"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/resource"
)

func NewDirect(addr string, opts ...func(cfg *Cfg)) (*grpc.ClientConn, error) {
	return getCfg(consts.Default, opts...).BuildDirect(addr)
}

func GetClient(service string, opts ...func(cfg *Cfg)) *Client {
	var fn = func(cfg *Cfg) {}
	if len(opts) > 0 {
		fn = opts[0]
	}
	return &Client{service: service, optFn: fn}
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

func (t *Client) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	var conn, err = t.Get()
	if err != nil {
		return xerror.Wrap(err)
	}
	return xerror.Wrap(conn.Invoke(ctx, method, args, reply, opts...))
}

func (t *Client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	var conn, err = t.Get()
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	var c, err1 = conn.NewStream(ctx, desc, method, opts...)
	return c, xerror.Wrap(err1)
}

func (t *Client) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return xerror.Wrap(t.conn.Close())
}

func (t *Client) CheckHealth(opts ...grpc.CallOption) (*grpc_health_v1.HealthCheckResponse, error) {
	c, err := t.Get()
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), t.cfg.Timeout)
	defer cancel()
	return grpc_health_v1.NewHealthClient(c).Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: t.service}, opts...)
}

func (t *Client) Watch(ctx context.Context, in *grpc_health_v1.HealthCheckRequest, opts ...grpc.CallOption) (grpc_health_v1.Health_WatchClient, error) {
	c, err := t.Get()
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return grpc_health_v1.NewHealthClient(c).Watch(ctx, in, opts...)
}

// Get new grpc Client
func (t *Client) Get() (_ grpc.ClientConnInterface, err error) {
	defer xerror.RespErr(&err)

	if t.conn != nil && t.conn.GetState() == connectivity.Ready {
		return t.conn, nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// 双检, 避免多次创建
	time.Sleep(time.Millisecond * 10)
	if t.conn != nil && t.conn.GetState() == connectivity.Ready {
		return t.conn, nil
	}

	t.cfg = getCfg(consts.Default)
	t.optFn(t.cfg)

	t.conn, err = t.cfg.Build(t.service)
	xerror.PanicF(err, "dial %s error", t.service)
	return t.conn, nil
}
