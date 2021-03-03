package grpclient

import (
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

func (t *client) New() (*grpc.ClientConn, error) {
	conn, err := dial(t.target, t.opts...)
	if err != nil {
		return nil, xerror.WrapF(err, "dial %s error\n", t.target)
	}
	return conn, nil
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

	mu.Lock()
	defer mu.Unlock()

	// 双检, 避免多次创建
	client = t.getClient()
	if client != nil {
		return client, nil
	}

	conn, err := t.New()
	if err != nil {
		return nil, err
	}

	clients.Store(t.service, conn)
	return conn, nil
}
