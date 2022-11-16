package grpcc

import (
	"context"
	"fmt"
	"sync"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/xerr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/clients/grpcc/grpcc_resolver"
	"github.com/pubgo/lava/core/projectinfo"
	"github.com/pubgo/lava/core/requestid"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/logging/logmiddleware"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/service"
)

func New(cfg *grpcc_config.Cfg, log *logging.Logger) Interface {
	cfg = merge.Copy(grpcc_config.DefaultCfg(), cfg).Unwrap()
	var c = &clientImpl{cfg: cfg, log: log, middlewares: []service.Middleware{
		logmiddleware.Middleware(log),
		requestid.Middleware(),
		projectinfo.Middleware(),
	}}
	if cfg.Client.Block {
		c.Get().Unwrap()
	}

	return c
}

type clientImpl struct {
	log         *logging.Logger
	cfg         *grpcc_config.Cfg
	mu          sync.Mutex
	conn        grpc.ClientConnInterface
	middlewares []service.Middleware
}

func (t *clientImpl) Middleware(mm ...service.Middleware) {
	t.middlewares = append(t.middlewares, mm...)
}

func (t *clientImpl) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) (err error) {
	defer recovery.Err(&err, func(err xerr.XErr) xerr.XErr {
		return err.WithMeta("method", method).WithMeta("input", args)
	})

	var conn = t.Get().Unwrap()
	assert.Must(conn.Invoke(ctx, method, args, reply, opts...))
	return
}

func (t *clientImpl) Healthy(ctx context.Context) error {
	conn := t.Get()
	if conn.IsErr() {
		return xerr.WrapF(conn.Err(), "get client failed, service=%s", t.cfg.Srv)
	}

	_, err := grpc_health_v1.NewHealthClient(conn.Unwrap()).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return xerr.WrapF(err, "service %s heath check failed", t.cfg.Srv)
	}
	return nil
}

func (t *clientImpl) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	var conn = t.Get()
	if conn.IsErr() {
		return nil, xerr.WrapF(conn.Err(), "get client failed, service=%s method=%s", t.cfg.Srv, method)
	}

	var c, err1 = conn.Unwrap().NewStream(ctx, desc, method, opts...)
	return c, xerr.Wrap(err1, method)
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

func createConn(cfg *grpcc_config.Cfg, log *logging.Logger, mm []service.Middleware) (grpc.ClientConnInterface, error) {
	// 创建grpc client
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Client.DialTimeout)
	defer cancel()

	addr := buildTarget(cfg)

	log.Info("grpc client init", zap.String(logkey.Service, cfg.Srv), zap.String("addr", addr))

	conn, err := grpc.DialContext(ctx, addr, append(cfg.Client.ToOpts(),
		grpc.WithChainUnaryInterceptor(unaryInterceptor(mm)),
		grpc.WithChainStreamInterceptor(streamInterceptor(mm)))...)
	if err != nil {
		return nil, xerr.WrapF(err, "grpc dial failed, target=>%s", addr)
	}

	log.Info("grpc client init ok", zap.String(logkey.Service, cfg.Srv), zap.String("addr", addr))
	return conn, nil
}
