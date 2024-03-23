package grpcc

import (
	"context"
	"fmt"
	"sync"

	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/vars"
	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/internal/middlewares/middleware_service_info"
	"github.com/pubgo/lava/lava"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/resolver"
)

type Params struct {
	Log       log.Logger
	Metric    metrics.Metric
	Resolvers []resolver.Builder
}

func New(cfg *grpcc_config.Cfg, p Params, middlewares ...lava.Middleware) Client {
	cfg = config.MergeR(grpcc_config.DefaultCfg(), cfg).Unwrap()
	cfg.Resolvers = p.Resolvers

	var defaultMiddlewares = lava.Middlewares{
		middleware_service_info.New(),
		middleware_metric.New(p.Metric),
		middleware_accesslog.New(p.Log.WithFields(log.Map{"service": cfg.Service.Name})),
		middleware_recovery.New(),
	}
	defaultMiddlewares = append(defaultMiddlewares, middlewares...)

	c := &clientImpl{
		cfg:         cfg,
		log:         p.Log,
		middlewares: defaultMiddlewares,
	}

	if cfg.Client.Block {
		_ = c.Get().Unwrap()
	}

	vars.RegisterValue(fmt.Sprintf("%s-grpc-client-config", cfg.Service.Name), cfg)
	return c
}

type clientImpl struct {
	log         log.Logger
	cfg         *grpcc_config.Cfg
	mu          sync.Mutex
	conn        grpc.ClientConnInterface
	middlewares []lava.Middleware
}

func (t *clientImpl) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) (err error) {
	defer recovery.Err(&err, func(err error) error {
		return errors.WrapTag(err, errors.T("method", method), errors.T("args", args))
	})

	conn := t.Get()
	if conn.IsErr() {
		return errors.Wrapf(conn.Err(), "failed to get grpc client, service=%s, method=%s", t.cfg.Service, method)
	}

	return conn.Unwrap().Invoke(ctx, method, args, reply, opts...)
}

func (t *clientImpl) Healthy(ctx context.Context) error {
	conn := t.Get()
	if conn.IsErr() {
		return errors.Wrapf(conn.Err(), "failed to get grpc client, service=%s, method=healthy", t.cfg.Service)
	}

	_, err := grpc_health_v1.NewHealthClient(conn.Unwrap()).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	return errors.Wrapf(err, "service %s heath check failed", t.cfg.Service)
}

func (t *clientImpl) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	conn := t.Get()
	if conn.IsErr() {
		return nil, errors.Wrapf(conn.Err(), "failed to get grpc client, service=%s, method=%s", t.cfg.Service, method)
	}

	c, err1 := conn.Unwrap().NewStream(ctx, desc, method, opts...)
	return c, errors.Wrap(err1, method)
}

// Get new grpc client
func (t *clientImpl) Get() (r result.Result[grpc.ClientConnInterface]) {
	defer recovery.Result(&r)

	if t.conn != nil {
		return r.WithVal(t.conn)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// 双检, 避免多次创建
	if t.conn != nil {
		return r.WithVal(t.conn)
	}

	conn, err := createConn(t.cfg, t.log, t.middlewares)
	if err != nil {
		return r.WithErr(err)
	}

	t.conn = conn
	return r.WithVal(t.conn)
}
