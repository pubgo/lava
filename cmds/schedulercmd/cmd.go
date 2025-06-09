package schedulercmd

import (
	"context"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/core/supervisor"
	"github.com/pubgo/lava/pkg/cmdutil"
	"github.com/pubgo/lava/servers/tasks"
)

func New(di dix.Container) *cli.Command {
	return &cli.Command{
		Name:  "scheduler",
		Usage: cmdutil.UsageDesc("crontab scheduler service %s(%s)", version.Project(), version.Version()),
		Action: func(ctx context.Context, command *cli.Command) error {
			s := dix.InjectMust(di, new(struct {
				Scheduler *scheduler.Scheduler
			}))

			srv := dix.InjectMust(di, tasks.New(s.Scheduler))
			return errors.WrapCaller(supervisor.Run(ctx, srv))
		},
	}
}
