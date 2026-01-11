package versioncmd

import (
	"context"
	"fmt"

	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/running"
	"github.com/pubgo/redant"

	"github.com/pubgo/lava/v2/pkg/cliutil"
)

func New() *redant.Command {
	return &redant.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   cliutil.UsageDesc("%s version info", version.Project()),
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			defer recovery.Exit()
			running.CheckVersion()

			fmt.Println("project:", version.Project())
			fmt.Println("version:", version.Version())
			fmt.Println("commit-id:", version.CommitID())
			fmt.Println("build-time:", version.BuildTime())
			fmt.Println("instance-id:", running.InstanceID)
			fmt.Println("system-info:", running.GetSysInfo())
			return nil
		},
	}
}
