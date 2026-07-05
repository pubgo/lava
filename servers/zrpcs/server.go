package zrpcs

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/vars"

	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/core/supervisor"
	"github.com/pubgo/lava/v2/pkg/middleware/accesslog"
	"github.com/pubgo/lava/v2/pkg/middleware/metric"
	"github.com/pubgo/lava/v2/pkg/middleware/recovery"
	"github.com/pubgo/lava/v2/pkg/middleware/serviceinfo"
	"github.com/pubgo/lava/v2/pkg/lava"
	"github.com/pubgo/lava/v2/pkg/zrpc"
)

type RegisterFunc func(*zrpc.Server) error

type Params struct {
	Registers   []RegisterFunc
	Middlewares []lava.Middleware
	Metric      metrics.Metric
	Log         log.Logger
	Conf        *Config
	// Started is closed after NATS subscriptions are registered and flushed.
	// Optional; useful in tests and startup orchestration.
	Started chan struct{}
}

func New(params Params) supervisor.Service { return newService(params) }

func newService(params Params) supervisor.Service {
	s := &serviceImpl{}
	s.init(params)
	return supervisor.NewService(s.String(), s.Serve)
}

type serviceImpl struct {
	log       log.Logger
	conf      *Config
	registers []RegisterFunc
	mw        []lava.Middleware
	started   chan struct{}
	nc        *nats.Conn
	srv       *zrpc.Server
}

func (s *serviceImpl) String() string { return "zrpc-server" }

func (s *serviceImpl) init(params Params) {
	s.conf = config.MergeR(defaultCfg(), params.Conf).Unwrap()
	s.log = params.Log.WithName(s.String())
	s.registers = params.Registers
	s.started = params.Started

	s.mw = make(lava.Middlewares, 0, 4+len(params.Middlewares))
	s.mw = append(s.mw,
		serviceinfo.New(),
		metric.New(params.Metric),
		accesslog.New(s.log),
		recovery.New(),
	)
	s.mw = append(s.mw, params.Middlewares...)

	vars.Register(vars.UniqueName("lava", "zrpc_server_info"), func() any {
		return map[string]any{"config": s.conf}
	})
}

func (s *serviceImpl) Serve(ctx context.Context) error {
	if err := s.start(); err != nil {
		return err
	}
	defer s.stop()

	<-ctx.Done()
	return nil
}

func (s *serviceImpl) start() error {
	nc, err := nats.Connect(s.conf.URL)
	if err != nil {
		return err
	}

	s.nc = nc
	s.srv = zrpc.NewServer(nc, s.mw...)
	for _, register := range s.registers {
		if err = register(s.srv); err != nil {
			s.stop()
			return err
		}
	}

	if err = s.nc.Flush(); err != nil {
		s.stop()
		return err
	}

	s.log.Info().Str("url", s.conf.URL).Int("registers", len(s.registers)).Msg("zrpc server started")
	if s.started != nil {
		close(s.started)
		s.started = nil
	}
	return nil
}

func (s *serviceImpl) stop() {
	if s.srv != nil {
		s.srv.Close()
		s.srv = nil
	}

	if s.nc != nil {
		s.log.Info().Str("url", s.conf.URL).Msg("zrpc server stopped")
		s.nc.Close()
		s.nc = nil
	}
}
