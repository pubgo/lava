package versioncmd

import (
	"fmt"

	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/pkg/cmdutil"
)

func New() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   cmdutil.UsageDesc("%s version info", version.Project()),
		Action: func(ctx *cli.Context) error {
			defer recovery.Exit()
			fmt.Println("project:", version.Project())
			fmt.Println("version:", version.Version())
			fmt.Println("commit-id:", version.CommitID())
			fmt.Println("build-time:", version.BuildTime())
			fmt.Println("instance-id:", running.InstanceID)
			return nil
		},
	}
}
