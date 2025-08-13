package supervisor

import (
	"context"
	"fmt"

	"github.com/pubgo/funk/errors"
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
}

func (s *serviceImpl) Name() string {
	return s.name
}

func (s *serviceImpl) String() string {
	return s.name
}

func (s *serviceImpl) Serve(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := s.fn(ctx)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("non context error, service=%s, err=%w", s.name, err)
	}
	return err
}
