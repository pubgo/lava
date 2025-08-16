package grpcservercmd

import (
	"context"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/pkg/cmdutil"
	"github.com/pubgo/lava/servers/tasks"
)

func New(di *dix.Dix) *cli.Command {
	return &cli.Command{
		Name:  "task",
		Usage: cmdutil.UsageDesc("async task service %s(%s)", version.Project(), version.Version()),
		Action: func(ctx context.Context, command *cli.Command) error {
			params := dix.Inject(di, tasks.Params{})
			manager := tasks.New(params)
			return manager.Run()
		},
	}
}
