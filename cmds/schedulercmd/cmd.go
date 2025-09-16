package schedulercmd

import (
	"context"

	"github.com/pubgo/dix/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/scheduler"
	"github.com/pubgo/lava/v2/core/scheduler/schedulerbuilder"
	"github.com/pubgo/lava/v2/core/scheduler/schedulerdebug"
	"github.com/pubgo/lava/v2/core/supervisor"
	"github.com/pubgo/lava/v2/pkg/cmdutil"
	"github.com/pubgo/lava/v2/servers/https"
)

func New(di *dix.Dix) *cli.Command {
	return &cli.Command{
		Name:  "scheduler",
		Usage: cmdutil.UsageDesc("crontab scheduler service %s(%s)", version.Project(), version.Version()),
		Action: func(ctx context.Context, command *cli.Command) error {
			di.Provide(schedulerbuilder.NewService)
			di.Provide(https.New)
			params := dix.Inject(di, new(struct {
				LC       lifecycle.Getter
				Services []supervisor.Service
				Manager  scheduler.JobManager
			}))

			schedulerdebug.Init(params.Manager)

			manager := supervisor.Default(params.LC)
			for _, svc := range params.Services {
				assert.Exit(manager.Add(svc))
			}

			return manager.Run(ctx)
		},
	}
}
