package grpcservercmd

import (
	"context"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/core/supervisor"
	"github.com/pubgo/lava/pkg/cmdutil"
	"github.com/pubgo/lava/servers/grpcs"
)

func New(di *dix.Dix) *cli.Command {
	return &cli.Command{
		Name:  "grpc",
		Usage: cmdutil.UsageDesc("grpc service %s(%s)", version.Project(), version.Version()),
		Action: func(ctx context.Context, command *cli.Command) error {
			dix.Provide(di, grpcs.New)
			m := dix.Inject(di, new(struct {
				Manager *supervisor.Manager
			}))
			return m.Manager.Run()
		},
	}
}
