package grpcc

import (
	"context"
	"fmt"
	"sync"

	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/xerr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/clients/grpcc/grpcc_resolver"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/merge"
	middleware2 "github.com/pubgo/lava/service"
)

var _ grpc.ClientConnInterface = (*Client)(nil)

func New(cfg *grpcc_config.Cfg, log *logging.Logger) *Client {
	cfg = merge.Copy(grpcc_config.DefaultCfg(), cfg).Unwrap()
	return &Client{cfg: cfg, log: log}
}

type Client struct {
	log  *logging.Logger
	cfg  *grpcc_config.Cfg
	mu   sync.Mutex
	conn grpc.ClientConnInterface
}

func (t *Client) createConn(cfg *grpcc_config.Cfg) (grpc.ClientConnInterface, error) {
	// 创建grpc client
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Client.DialTimeout)
	defer cancel()

	var middlewares []middleware2.Middleware
	for _, m := range t.cfg.Middlewares {
		middlewares = append(middlewares, m)
	}

	addr := t.buildTarget(cfg)
	conn, err := grpc.DialContext(ctx, addr, append(cfg.Client.ToOpts(),
		grpc.WithChainUnaryInterceptor(unaryInterceptor(middlewares)),
		grpc.WithChainStreamInterceptor(streamInterceptor(middlewares)))...)

	logging.L().Info("grpc client init", zap.String(logkey.Service, cfg.Srv))
	logutil.Pretty(err)
	return conn, xerr.WrapF(err, "grpc dial failed, target=>%s", addr)
}

func (t *Client) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	var conn, err = t.Get()
	if err != nil {
		return xerr.Wrap(err, method, args)
	}

	return xerr.Wrap(conn.Invoke(ctx, method, args, reply, opts...), method)
}

func (t *Client) Check(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) error {
	conn, err := t.Get()
	if err != nil {
		return xerr.WrapF(err, "service %s heath check failed", t.cfg.Srv)
	}

	_, err = grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return xerr.WrapF(err, "service %s heath check failed", t.cfg.Srv)
	}
	return nil
}

func (t *Client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	var conn, err = t.Get()
	if err != nil {
		return nil, xerr.Wrap(err, method)
	}

	var c, err1 = conn.NewStream(ctx, desc, method, opts...)
	return c, xerr.Wrap(err1, method)
}

// Get new grpc Client
func (t *Client) Get() (_ grpc.ClientConnInterface, gErr error) {
	defer recovery.Recovery(func(err xerr.XErr) {
		gErr = err

		logutil.Pretty(t)
		err.DebugPrint()
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

	conn, err := t.createConn(t.cfg)
	if err != nil {
		return nil, err
	}

	t.conn = conn
	return t.conn, nil
}

func (t *Client) buildTarget(cfg *grpcc_config.Cfg) string {
	var addr = cfg.Srv
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
