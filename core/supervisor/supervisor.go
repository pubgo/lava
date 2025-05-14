package supervisor

import (
	"context"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/running"
	"github.com/thejerf/suture/v4"
)

type Service = suture.Service
type Supervisor = suture.Supervisor

func New() *Supervisor {
	return suture.NewSimple(running.Project)
}

func Run(ctx context.Context, services ...suture.Service) error {
	if len(services) == 0 {
		return nil
	}

	manager := suture.NewSimple(running.Project)
	for _, service := range services {
		manager.Add(service)
	}
	return errors.WrapCaller(manager.Serve(ctx))
}
