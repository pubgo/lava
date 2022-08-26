package grpcservercmd

import (
	"fmt"
	"github.com/pubgo/lava/internal/service/grpcs"

	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/version"
)

func Cmd() *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: fmt.Sprintf("%s service", version.Project()),
		Action: func(ctx *cli.Context) error {
			defer recovery.Exit()
			var srv = grpcs.New()
			srv.Start()
			signal.Wait()
			srv.Stop()
			return nil
		},
	}
}
