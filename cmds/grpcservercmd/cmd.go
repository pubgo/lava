package grpcservercmd

import (
	"fmt"
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/service/grpcs"

	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/version"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "grpc",
		Usage: fmt.Sprintf("%s service", version.Project()),
		Action: func(ctx *cli.Context) error {
			defer recovery.Exit()
			srv := di.Inject(grpcs.New())
			srv.Start()
			signal.Wait()
			srv.Stop()
			return nil
		},
	}
}
