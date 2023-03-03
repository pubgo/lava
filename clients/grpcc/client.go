package grpcc

import (
	"context"
	"fmt"
	"sync"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/merge"
	"github.com/pubgo/funk/metric"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/clients/grpcc/grpcc_resolver"
	"github.com/pubgo/lava/internal/middlewares/middleware_log"
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/logging/logkey"
)

func New(cfg *grpcc_config.Cfg, log log.Logger, m metric.Metric) Interface {
	cfg = merge.Copy(generic.Ptr(grpcc_config.DefaultCfg()), cfg).Unwrap()
	var c = &clientImpl{
		cfg: cfg,
		log: log,
		middlewares: []lava.Middleware{
			middleware_metric.New(m),
			middleware_log.New(log),
			middleware_recovery.New(),
		},
	}

	if cfg.Client.Block {
		c.Get().Unwrap()
	}

	return c
}

type clientImpl struct {
	log         log.Logger
	cfg         *grpcc_config.Cfg
	mu          sync.Mutex
	conn        grpc.ClientConnInterface
	middlewares []lava.Middleware
}

func (t *clientImpl) Middleware(mm ...lava.Middleware) {
	t.middlewares = append(t.middlewares, mm...)
}

func (t *clientImpl) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) (err error) {
	defer recovery.Err(&err, func(err *errors.Event) {
		err.Str("method", method).Any("input", args)
	})

	var conn = t.Get().Unwrap()
	assert.Must(conn.Invoke(ctx, method, args, reply, opts...))
	return
}

func (t *clientImpl) Healthy(ctx context.Context) error {
	conn := t.Get()
	if conn.IsErr() {
		return errors.Wrapf(conn.Err(), "get client failed, service=%s", t.cfg.Srv)
	}

	_, err := grpc_health_v1.NewHealthClient(conn.Unwrap()).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return errors.Wrapf(err, "service %s heath check failed", t.cfg.Srv)
	}
	return nil
}

func (t *clientImpl) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	var conn = t.Get()
	if conn.IsErr() {
		return nil, errors.Wrapf(conn.Err(), "get client failed, service=%s method=%s", t.cfg.Srv, method)
	}

	var c, err1 = conn.Unwrap().NewStream(ctx, desc, method, opts...)
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

func buildTarget(cfg *grpcc_config.Cfg) string {
	var addr = cfg.Addr
	var scheme = grpcc_resolver.DirectScheme
	if cfg.Scheme != "" {
		scheme = cfg.Scheme
	}

	switch scheme {
	case grpcc_resolver.DiscovScheme:
		return grpcc_resolver.BuildDiscovTarget(addr)
	case grpcc_resolver.DirectScheme:
		return grpcc_resolver.BuildDirectTarget(addr)
	case grpcc_resolver.K8sScheme, grpcc_resolver.DnsScheme:
		return fmt.Sprintf("dns:///%s", addr)
	default:
		return addr
	}
}

func createConn(cfg *grpcc_config.Cfg, log log.Logger, mm []lava.Middleware) (grpc.ClientConnInterface, error) {
	// 创建grpc client
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Client.DialTimeout)
	defer cancel()

	addr := buildTarget(cfg)

	var ee = log.Info().
		Str(logkey.Service, cfg.Srv).
		Str("addr", addr)
	ee.Msg("grpc client init")

	conn, err := grpc.DialContext(ctx, addr, append(
		cfg.Client.ToOpts(),
		grpc.WithChainUnaryInterceptor(unaryInterceptor(mm)),
		grpc.WithChainStreamInterceptor(streamInterceptor(mm)))...)
	if err != nil {
		return nil, errors.Wrapf(err, "grpc dial failed, target=>%s", addr)
	}

	ee.Msg("grpc client init ok")
	return conn, nil
}
