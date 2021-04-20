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
var clients sync.Map

func NewClient(service string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return GetDefaultCfg().Build(service, opts...)
}

func GetClient(service string, optFns ...func(service string) []grpc.DialOption) *client {
	var fn = defaultDialOption
	if len(optFns) > 0 {
		fn = optFns[0]
	}

	return &client{
		service: service,
		optFn:   fn,
	}
}

var _ GrpcClient = (*client)(nil)

type client struct {
	service string
	optFn   func(service string) []grpc.DialOption
}

func (t *client) getClient() *grpc.ClientConn {
	if val, ok := clients.Load(t.service); ok {
		if val.(*grpc.ClientConn).GetState() == connectivity.Ready {
			return val.(*grpc.ClientConn)
		}
	}
	return nil
}

// ClientProtocol impl.
func (t *client) Ping() error {
	c, err := t.Get()
	if err != nil {
		return xerror.Wrap(err)
	}

	client := grpc_health_v1.NewHealthClient(c)
	var ctx = context.Background()
	_, err = client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: t.service})
	return err
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
