package versioncmd

import (
	"context"
	"fmt"

	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/running"
	"github.com/pubgo/lava/v2/pkg/cliutil"
	"github.com/urfave/cli/v3"
)

func New() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   cliutil.UsageDesc("%s version info", version.Project()),
		Action: func(ctx context.Context, command *cli.Command) error {
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
