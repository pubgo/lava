package versioncmd

import (
	"fmt"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/core/runmode"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/pkg/cmdx"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/version"
)

func New() *cli.Command {
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
			fmt.Println("version:", version.Version())
			fmt.Println("commit_id:", version.CommitID())
			fmt.Println("project:", version.Project())
			fmt.Println("device_id:", runmode.DeviceID)
			return nil
		},
	}
}
