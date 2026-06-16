package zrpcc

import (
	"context"
	"sync"
	"time"

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
	mu          sync.RWMutex
	closed      bool
	nc          *nats.Conn
	rt          *zrpc.Client
}

func (c *clientImpl) connectLocked() (*nats.Conn, *zrpc.Client, error) {
	if c.closed {
		return nil, nil, errors.New("client is closed")
	}

	if c.rt == nil && c.nc != nil {
		return nil, nil, errors.New("client is closed")
	}

	if c.nc != nil {
		return c.nc, c.rt, nil
	}

	nc, err := nats.Connect(c.cfg.URL)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to connect zrpc server, url=%s", c.cfg.URL)
	}

	c.nc = nc
	c.rt = zrpc.NewClient(nc, c.middlewares...)
	c.log.Info().Str("url", c.cfg.URL).Msg("zrpc client connected")
	return c.nc, c.rt, nil
}

func (c *clientImpl) withRuntime(fn func(*zrpc.Client, time.Duration) error) error {
	c.mu.RLock()
	if c.rt != nil {
		defer c.mu.RUnlock()
		return fn(c.rt, c.cfg.Timeout)
	}
	c.mu.RUnlock()

	c.mu.Lock()
	_, rt, err := c.connectLocked()
	timeout := c.cfg.Timeout
	if err != nil {
		c.mu.Unlock()
		return err
	}
	c.mu.Unlock()

	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.rt == nil || c.rt != rt {
		return errors.New("client is closed")
	}

	return fn(rt, timeout)
}

func (c *clientImpl) withConn(fn func(*nats.Conn) error) error {
	c.mu.RLock()
	if c.nc != nil {
		defer c.mu.RUnlock()
		return fn(c.nc)
	}
	c.mu.RUnlock()

	c.mu.Lock()
	nc, _, err := c.connectLocked()
	if err != nil {
		c.mu.Unlock()
		return err
	}
	c.mu.Unlock()

	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.nc == nil || c.nc != nc {
		return errors.New("client is closed")
	}

	return fn(nc)
}

func (c *clientImpl) Conn() (*nats.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	nc, _, err := c.connectLocked()
	return nc, err
}

func (c *clientImpl) CallUnary(ctx context.Context, subject string, req, resp proto.Message) error {
	return c.withRuntime(func(rt *zrpc.Client, timeout time.Duration) error {
		return rt.CallUnary(ctx, subject, timeout, req, resp)
	})
}

func (c *clientImpl) OpenStream(ctx context.Context, subject string) (*zrpc.ClientStream, error) {
	var stream *zrpc.ClientStream
	err := c.withRuntime(func(rt *zrpc.Client, timeout time.Duration) error {
		var openErr error
		stream, openErr = rt.OpenStream(ctx, subject, timeout)
		return openErr
	})
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (c *clientImpl) Healthy(ctx context.Context) error {
	return c.withConn(func(nc *nats.Conn) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		flushTimeout := 5 * time.Second
		if deadline, ok := ctx.Deadline(); ok {
			flushTimeout = time.Until(deadline)
			if flushTimeout <= 0 {
				return context.DeadlineExceeded
			}
		}

		if err := nc.FlushTimeout(flushTimeout); err != nil {
			return errors.Wrap(err, "failed to flush nats connection")
		}

		if err := nc.LastError(); err != nil {
			return errors.Wrap(err, "nats connection unhealthy")
		}

		return nil
	})
}

func (c *clientImpl) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
	if c.nc == nil {
		return nil
	}

	c.log.Info().Str("url", c.cfg.URL).Msg("zrpc client closed")
	c.nc.Close()
	c.nc = nil
	c.rt = nil
	return nil
}
