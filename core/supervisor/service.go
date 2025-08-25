package supervisor

import (
	"context"
	"expvar"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/vars"
	"github.com/pubgo/lava/core/debug"
)

func NewService(name string, fn func(ctx context.Context) error) Service {
	return (&serviceImpl{name: name, fn: fn}).initMetric()
}

var _ Service = &serviceImpl{}

type serviceImpl struct {
	name string
	err  error
	fn   func(ctx context.Context) error

	metric *expvar.Map
}

func (s *serviceImpl) Metrics() *expvar.Map {
	return s.metric
}

func (s *serviceImpl) initMetric() *serviceImpl {
	metric := vars.Map(fmt.Sprintf("supervisor-%s", s.name))
	metric.Set(s.name, vars.Any(s.String()))
	metric.Set(s.name+".error", vars.Value(func() interface{} {
		if s.err == nil {
			return nil
		}
		return s.err.Error()
	}))
	metric.Add("restart", 0)
	s.metric = metric

	debug.Route("/supervisor", func(router fiber.Router) {
		router.Get("services", func(ctx *fiber.Ctx) error {

			return nil
		})
	})

	return s
}

func (s *serviceImpl) Error() error {
	return s.err
}

func (s *serviceImpl) Name() string {
	return s.name
}

func (s *serviceImpl) String() string {
	return s.name
}

func (s *serviceImpl) Serve(ctx context.Context) (gErr error) {
	now := time.Now()
	defer func() {
		if gErr != nil {
			s.err = gErr
		}

		if s.err != nil {
			s.metric.Set("error", vars.Any(s.err))
		}

		s.metric.Set("start_time", vars.Any(now.UTC().String()))
		s.metric.Set("online_duration", vars.Any(time.Since(now).String()))
		log.Info(ctx).
			Str("service", s.name).
			Str("metric", s.metric.String()).
			Msg("stop service")
	}()
	defer recovery.Err(&gErr)

	s.err = nil
	s.metric.Add("restart", 1)
	log.Info(ctx).Str("service", s.name).Msg("start service")
	err := s.fn(ctx)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("non-context error, service=%s meta=%s err=%w", s.name, s.metric.String(), err)
	}
	return err
}
