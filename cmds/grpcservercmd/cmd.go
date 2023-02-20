package grpcservercmd

import (
	"fmt"

	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/servers/grpcs"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "grpc",
		Usage: fmt.Sprintf("%s grpc service", version.Project()),
		Action: func(ctx *cli.Context) error {
			defer recovery.Exit()
			srv := di.Inject(grpcs.New())
			srv.Run()
			return nil
		},
	}
}
