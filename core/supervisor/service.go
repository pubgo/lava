package supervisor

import (
	"context"
	"expvar"
	"fmt"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/vars"
)

func NewService(name string, fn func(ctx context.Context) error) Service {
	srv := &serviceImpl{name: name, fn: fn}
	return srv.initMetric()
}

var _ Service = &serviceImpl{}

type serviceImpl struct {
	name string
	err  error
	fn   func(ctx context.Context) error

	metric *expvar.Map
}

func (s *serviceImpl) initMetric() *serviceImpl {
	metric := new(expvar.Map).Init()
	metric.Set(s.name, s)
	metric.Set(s.name+".error", vars.Value(func() interface{} { return s.err }))
	s.metric = metric
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
	defer recovery.Recovery(func(err error) {
		s.err = err
		gErr = err
	})

	s.err = nil
	s.metric.Add("restart", 1)
	err := s.fn(ctx)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("non-context error, service=%s meta=%s err=%w", s.name, s.metric.String(), err)
	}
	return err
}
