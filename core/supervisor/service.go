package supervisor

import (
	"context"
	"expvar"
	"fmt"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/recovery"
)

func NewService(name string, fn func(ctx context.Context) error) Service {
	return &serviceImpl{
		name: name,
		fn:   fn,
	}
}

var _ Service = &serviceImpl{}

type serviceImpl struct {
	name string
	err  error
	fn   func(ctx context.Context) error

	restartCounter expvar.Int
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
	s.restartCounter.Add(1)
	err := s.fn(ctx)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("non context error, service=%s, err=%w", s.name, err)
	}
	return err
}
