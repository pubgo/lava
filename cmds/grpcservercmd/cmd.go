package grpcservercmd

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/pkg/cmdutil"
	"github.com/pubgo/lava/servers/grpcs"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "grpc",
		Usage: cmdutil.UsageDesc("%s grpc service", version.Project()),
		Action: func(ctx *cli.Context) error {
			defer recovery.Exit()
			srv := di.Inject(grpcs.New())
			srv.Run()
			return nil
		},
	}
}
