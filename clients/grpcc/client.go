package grpcc

import (
	"context"
	"sync"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/clients/grpcc/endpoint"
	"github.com/pubgo/lava/clients/grpcc/resolver"
)

func New(service string, opts ...func(cfg *Cfg)) *Client {
	return NewClient(service, DefaultCfg(opts...))
}

// NewClient build grpc client
func NewClient(service string, cfg Cfg) *Client {
	var name = service
	switch cfg.buildScheme {
	case resolver.DiscovScheme:
		service = resolver.BuildDiscovTarget(service, cfg.registry)
	case resolver.DirectScheme:
		service = resolver.BuildDirectTarget(service)
	default:
		service, name = endpoint.Interpret(service)
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
	defer xerror.RespErr(&gErr)

	if t.conn != nil {
		return t.conn, nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// 双检, 避免多次创建
	if t.conn != nil {
		return t.conn, nil
	}

	// 创建grpc client
	ctx, cancel := context.WithTimeout(context.Background(), t.cfg.DialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, t.addr, append(t.cfg.ToOpts(), t.cfg.DialOptions...)...)
	xerror.PanicF(err, "DialContext error, target:%s\n", t.addr)
	t.conn = conn
	return t.conn, nil
}
