package initCmd

import (
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"
)

var protoCfg string

func Cmd() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "project init",
		Flags: typex.Flags{
			&cli.StringFlag{
				Name:        "protobuf",
				Usage:       "protobuf config path",
				Value:       protoCfg,
				Destination: &protoCfg,
			},
		},
		Before: func(ctx *cli.Context) error {
			defer xerror.RecoverAndExit()

			return nil
		},
		Subcommands: cli.Commands{},
	}
}
