package schedulerbuilder

import (
	"github.com/pubgo/lava/v2/core/scheduler"
	"github.com/pubgo/lava/v2/core/supervisor"
)

type ResponseParams struct {
	Service supervisor.Service
	Manager scheduler.JobManager
}

func NewService(params scheduler.Params) (ResponseParams, error) {
	s, err := scheduler.New(
		params.M,
		params.Log,
		params.Metric,
		params.Configs,
		params.Routers,
		params.Executors,
	)
	if err != nil {
		return ResponseParams{}, err
	}

	return ResponseParams{
		Service: supervisor.NewService(scheduler.Name, s.Serve),
		Manager: s,
	}, nil
}
