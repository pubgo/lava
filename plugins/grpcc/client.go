package grpcc

import (
	"context"
	"sync"

	"github.com/pubgo/lug/consts"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var mux sync.Mutex

var _ Client = (*client)(nil)

type client struct {
	service string
	optFn   func(service string) []grpc.DialOption
}

func (t *client) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest, opts ...grpc.CallOption) (*grpc_health_v1.HealthCheckResponse, error) {
	c, err := t.Get()
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	return grpc_health_v1.NewHealthClient(c).Check(ctx, in, opts...)
}

func (t *client) Watch(ctx context.Context, in *grpc_health_v1.HealthCheckRequest, opts ...grpc.CallOption) (grpc_health_v1.Health_WatchClient, error) {
	c, err := t.Get()
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	return grpc_health_v1.NewHealthClient(c).Watch(ctx, in, opts...)
}

func (t *client) getClient() *grpc.ClientConn {
	if val, ok := clients.Load(t.service); ok {
		if val.(*grpc.ClientConn).GetState() == connectivity.Ready {
			return val.(*grpc.ClientConn)
		}
	}
	return nil
}

// Get new grpc client
func (t *client) Get() (*grpc.ClientConn, error) {
	var client = t.getClient()
	if client != nil {
		return client, nil
	}

	mux.Lock()
	defer mux.Unlock()

	// 双检, 避免多次创建
	client = t.getClient()
	if client != nil {
		return client, nil
	}

	var cfg = GetCfg(consts.Default)
	conn, err := cfg.Build(t.service, t.optFn(t.service)...)
	if err != nil {
		return nil, xerror.WrapF(err, "dial %s error\n", t.service)
	}

	clients.Store(t.service, conn)
	return conn, nil
}
