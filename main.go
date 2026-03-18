package main

import (
	"context"
	"fmt"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/redant"

	"github.com/pubgo/lava/v2/cmds/curlcmd"
	"github.com/pubgo/lava/v2/cmds/devproxycmd"
	"github.com/pubgo/lava/v2/cmds/fileservercmd"
	"github.com/pubgo/lava/v2/cmds/tunnelcmd"
	"github.com/pubgo/lava/v2/cmds/watchcmd"
	"github.com/pubgo/lava/v2/core/lavabuilder"
	"github.com/pubgo/lava/v2/core/signals"
	"github.com/pubgo/lava/v2/pkg/cliutil"
)

func main() {
	defer recovery.Exit()

	// 创建主命令
	di := lavabuilder.New()
	app := &redant.Command{
		Use:   "lava",
		Short: cliutil.UsageDesc("%s service", version.Project()),
		Long:  "Lava is a microservice integration framework",
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			fmt.Println("Usage: lava [command] [arguments]")
			fmt.Println("Available commands:")
			fmt.Println("  watch     Watch files for changes and run commands automatically")
			fmt.Println("  curl      Make HTTP requests to gRPC services")
			fmt.Println("  tunnel    Tunnel gateway commands")
			fmt.Println("  devproxy  Local development proxy tool")
			return nil
		},
		// 添加子命令
		Children: []*redant.Command{
			watchcmd.New(),
			curlcmd.New(),
			tunnelcmd.New(di),
			fileservercmd.New(),
			devproxycmd.New(di),
		},
	}

	// 运行命令
	assert.Exit(app.Run(signals.Context()))
}
