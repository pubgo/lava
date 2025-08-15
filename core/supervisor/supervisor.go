package supervisor

import (
	"context"
	"fmt"

	"github.com/thejerf/suture/v4"
)

type Service interface {
	Name() string
	Error() error
	fmt.Stringer
	suture.Service
}

type Supervisor = suture.Supervisor
type Spec = suture.Spec

type serviceFn func(ctx context.Context) error

func (fn serviceFn) Serve(ctx context.Context) error {
	return fn(ctx)
}
