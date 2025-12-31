package grpcc

import (
	"context"
	"fmt"
	"sync"

	"github.com/pubgo/lava/v2/clients/grpcc/grpccconfig"
	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/lava"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/resolver"

	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/funk/v2/vars"
)

type Params struct {
	Log       log.Logger
	Metric    metrics.Metric
	Resolvers []resolver.Builder
}

func New(cfg *grpccconfig.Cfg, p Params, middlewares ...lava.Middleware) Client {
	cfg = config.MergeR(grpccconfig.DefaultCfg(), cfg).Unwrap()
	cfg.Resolvers = p.Resolvers

	c := &clientImpl{
		cfg:         cfg,
		log:         p.Log,
		middlewares: middlewares,
	}

	vars.Register(fmt.Sprintf("%s-grpc-client-config", cfg.Service.Name), func() any { return cfg })
	return c
}

type clientImpl struct {
	log         log.Logger
	cfg         *grpccconfig.Cfg
	mu          sync.Mutex
	conn        grpc.ClientConnInterface
	middlewares []lava.Middleware
}

func (t *clientImpl) Invoke(ctx context.Context, method string, args, reply any, opts ...grpc.CallOption) (err error) {
	defer recovery.Err(&err, func(err error) error {
		return errors.WrapTags(err, errors.Tags{"method": method, "args": args})
	})

	conn := t.Get().
		MapErr(func(err error) error {
			return errors.Wrapf(err, "failed to get grpc client, service=%s, method=%s", t.cfg.Service, method)
		}).
		CallIfOK(func(val grpc.ClientConnInterface) error {
			return val.Invoke(ctx, method, args, reply, opts...)
		})

	return conn.Err()
}

func (t *clientImpl) Healthy(ctx context.Context) error {
	return t.Get().
		MapErr(func(err error) error {
			return errors.Wrapf(err, "failed to get grpc client, service=%s, method=healthy", t.cfg.Service)
		}).
		CallIfOK(func(val grpc.ClientConnInterface) error {
			_, err := grpc_health_v1.NewHealthClient(val).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
			return errors.Wrapf(err, "service %s heath check failed", t.cfg.Service)
		}).
		Err()
}

func (t *clientImpl) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	conn := t.Get().MapErr(func(err error) error {
		return errors.Wrapf(err, "failed to get grpc client, service=%s, method=%s", t.cfg.Service, method)
	})
	return result.FlatMapTo(conn, func(val grpc.ClientConnInterface) (r result.Result[grpc.ClientStream]) {
		return result.Wrap(val.NewStream(ctx, desc, method, opts...)).
			MapErr(func(err error) error {
				return errors.Wrapf(err, "service %s:%s new stream failed", t.cfg.Service, method)
			})
	}).UnwrapErr()
}

// Get new grpc client
func (t *clientImpl) Get() (r result.Result[grpc.ClientConnInterface]) {
	defer result.Recovery(&r)

	if t.conn != nil {
		return r.WithValue(t.conn)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// 双检, 避免多次创建
	if t.conn != nil {
		return r.WithValue(t.conn)
	}

	conn := createConn(t.cfg, t.log, t.middlewares).UnwrapOrThrow(&r)
	if r.IsErr() {
		return r
	}

	t.conn = conn
	return r.WithValue(t.conn)
}
