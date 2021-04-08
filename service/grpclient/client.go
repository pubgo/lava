package grpclient

import (
	"context"
	"google.golang.org/grpc/health/grpc_health_v1"
	"sync"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var mu sync.Mutex
var clients sync.Map

func Client(service string, opts ...grpc.DialOption) *client {
	return &client{
		service: service,
		target:  buildTarget(service),
		opts:    append(defaultDialOpts, opts...),
	}
}

type client struct {
	service string
	target  string
	opts    []grpc.DialOption
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

	mu.Lock()
	defer mu.Unlock()

	// 双检, 避免多次创建
	client = t.getClient()
	if client != nil {
		return client, nil
	}

	conn, err := dial(t.target, t.opts...)
	if err != nil {
		return nil, xerror.WrapF(err, "dial %s error\n", t.service)
	}

	clients.Store(t.service, conn)
	return conn, nil
}
