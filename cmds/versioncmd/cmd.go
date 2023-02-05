package versioncmd

import (
	"fmt"

	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/core/runmode"
)

func New() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "show the project version information",
		Action: func(ctx *cli.Context) error {
			defer recovery.Exit()
			fmt.Println("project:", version.Project())
			fmt.Println("version:", version.Version())
			fmt.Println("commit-id:", version.CommitID())
			fmt.Println("device-id:", runmode.DeviceID)
			fmt.Println("instance-id:", runmode.InstanceID)
			return nil
		},
	}
}
