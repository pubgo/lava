package versioncmd

import (
	"fmt"

	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/internal/pkg/cmdx"
	"github.com/pubgo/lava/internal/pkg/typex"
	"github.com/pubgo/lava/version"
)

func Cmd() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: typex.StrOf("v"),
		Usage:   "show the project version information",
		Description: cmdx.ExampleFmt(
			"lava version",
			"lava version json",
			"lava version t"),
		Action: func(ctx *cli.Context) error {
			defer recovery.Exit()
			fmt.Println(version.Domain())
			fmt.Println(version.Version())
			fmt.Println(version.CommitID())
			fmt.Println(version.Project())
			return nil
		},
	}
}
