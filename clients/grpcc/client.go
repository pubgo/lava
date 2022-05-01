package grpcc

import (
	"context"
	"sync"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/runtime"

	// 加载mdns注册中心
	_ "github.com/pubgo/lava/core/registry/registry_driver/mdns"

	// 加载grpcLog
	_ "github.com/pubgo/lava/logging/log_ext/grpclog"
)

var _ grpc.ClientConnInterface = (*Client)(nil)

func NewClient(srv string, opts ...Option) *Client {
	var cli = &Client{srv: srv, cfg: grpcc_config.DefaultCfg()}
	for i := range opts {
		opts[i](cli)
	}

	xerror.Assert(cli.dial == nil, "[dial] is nil")
	return cli
}

type Client struct {
	dial func(addr string, cfg grpcc_config.Cfg) (grpc.ClientConnInterface, error)
	cfg  grpcc_config.Cfg
	mu   sync.Mutex
	conn grpc.ClientConnInterface
	srv  string
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

		if !runtime.IsProd() {
			logutil.Pretty(t)
			err.Debug()
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

	var cfg = t.cfg
	var cfgMap = make(map[string]*grpcc_config.Cfg)
	xerror.Panic(config.Decode(grpcc_config.Name, &cfgMap))
	if cfgMap[consts.KeyDefault] != nil {
		xerror.Panic(merge.Copy(&cfg, cfgMap[consts.KeyDefault]))
	}
	if cfgMap[t.srv] != nil {
		xerror.Panic(merge.Copy(&cfg, cfgMap[t.srv]))
	}

	conn, err := t.dial(t.srv, cfg)
	if err != nil {
		return nil, err
	}

	t.conn = conn
	return t.conn, nil
}
