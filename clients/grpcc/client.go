package grpcc

import (
	"context"
	"sync"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/core/logging/logutil"
	"github.com/pubgo/lava/runtime"

	// 加载mdns注册中心
	_ "github.com/pubgo/lava/core/registry/registry_driver/mdns"

	// 加载grpcLog
	_ "github.com/pubgo/lava/core/logging/log_ext/grpclog"
)

var _ grpc.ClientConnInterface = (*Client)(nil)

func NewClient(srv string) *Client {
	return &Client{srv: srv}
}

type Client struct {
	dial func(addr string, cfg *grpcc_config.Cfg, plugins ...string) (*grpc.ClientConn, error)
	cfg  *grpcc_config.Cfg
	mu   sync.Mutex
	conn *grpc.ClientConn

	srv        string
	plugins    []string
	beforeDial func()
	afterDial  func()
}

func (t *Client) Plugin(plugins ...string) {
	t.plugins = append(t.plugins, plugins...)
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

	var addr = t.buildTarget(t.name)

	if t.beforeDial != nil {
		t.beforeDial()
	}

	conn, err := t.dial(addr, t.cfg, t.plugins...)
	if err != nil {
		return nil, err
	}

	if t.afterDial != nil {
		t.afterDial()
	}

	t.conn = conn
	return t.conn, nil
}
