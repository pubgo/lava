package zrpcc

import (
	"context"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/log"
	"google.golang.org/protobuf/proto"

	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_serviceinfo"
	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/zrpc"
)

type Params struct {
	Log    log.Logger
	Metric metrics.Metric
}

func New(cfg *Config, p Params, middlewares ...lava.Middleware) Client {
	cfg = config.MergeR(DefaultCfg(), cfg).Unwrap()

	mm := make(lava.Middlewares, 0, 4+len(middlewares))
	mm = append(mm,
		middleware_serviceinfo.New(),
		middleware_metric.New(p.Metric),
		middleware_accesslog.New(p.Log.WithFields(log.Fields{"service": Name})),
		middleware_recovery.New(),
	)
	mm = append(mm, middlewares...)

	return &clientImpl{
		cfg:         cfg,
		log:         p.Log.WithName(Name),
		middlewares: mm,
	}
}

type clientImpl struct {
	cfg         *Config
	log         log.Logger
	middlewares []lava.Middleware
	mu          sync.Mutex
	nc          *nats.Conn
	rt          *zrpc.Client
}

func (c *clientImpl) Conn() (*nats.Conn, error) {
	if c.nc != nil {
		return c.nc, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.nc != nil {
		return c.nc, nil
	}

	nc, err := nats.Connect(c.cfg.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect zrpc server, url=%s", c.cfg.URL)
	}

	c.nc = nc
	c.rt = zrpc.NewClient(nc, c.middlewares...)
	return c.nc, nil
}

func (c *clientImpl) CallUnary(ctx context.Context, subject string, req, resp proto.Message) error {
	if _, err := c.Conn(); err != nil {
		return err
	}

	return c.rt.CallUnary(ctx, subject, c.cfg.Timeout, req, resp)
}

func (c *clientImpl) OpenStream(ctx context.Context, subject string) (*zrpc.ClientStream, error) {
	if _, err := c.Conn(); err != nil {
		return nil, err
	}

	return c.rt.OpenStream(ctx, subject, c.cfg.Timeout)
}

func (c *clientImpl) Healthy(ctx context.Context) error {
	nc, err := c.Conn()
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err = nc.Flush(); err != nil {
		return errors.Wrap(err, "failed to flush nats connection")
	}

	if err = nc.LastError(); err != nil {
		return errors.Wrap(err, "nats connection unhealthy")
	}

	return nil
}

func (c *clientImpl) Close() error {
	if c.nc == nil {
		return nil
	}

	c.nc.Close()
	c.nc = nil
	c.rt = nil
	return nil
}
