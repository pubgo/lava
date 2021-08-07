package grpcc

import (
	"context"
	"sync"
	"time"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pubgo/lug/consts"
)

var _ Client = (*client)(nil)

type client struct {
	mu      sync.Mutex
	optFn   func(cfg *Cfg)
	service string
	conn    *grpc.ClientConn
	cfg     *Cfg
}

func (t *client) Close() error {
	t.mu.Lock()
	var conn = t.conn
	t.mu.Unlock()

	return xerror.Wrap(conn.Close())
}

func (t *client) Check(opts ...grpc.CallOption) (*grpc_health_v1.HealthCheckResponse, error) {
	c, err := t.Get()
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), t.cfg.Timeout)
	defer cancel()
	return grpc_health_v1.NewHealthClient(c).Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: t.service}, opts...)
}

func (t *client) Watch(ctx context.Context, in *grpc_health_v1.HealthCheckRequest, opts ...grpc.CallOption) (grpc_health_v1.Health_WatchClient, error) {
	c, err := t.Get()
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return grpc_health_v1.NewHealthClient(c).Watch(ctx, in, opts...)
}

// Get new grpc client
func (t *client) Get() (_ *grpc.ClientConn, err error) {
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

	t.cfg = GetCfg(consts.Default)
	t.optFn(t.cfg)

	t.conn, err = t.cfg.Build(t.service)
	xerror.PanicF(err, "dial %s error", t.service)
	//xerror.PanicErr(t.Check(nil))
	return t.conn, nil
}
