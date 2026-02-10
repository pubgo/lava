package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/redant"

	"github.com/pubgo/lava/v2/cmds/curlcmd"
	"github.com/pubgo/lava/v2/cmds/watchcmd"
	"github.com/pubgo/lava/v2/pkg/cliutil"
)

func main() {
	defer recovery.Exit(func(err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return err
	})

	// 创建主命令
	app := &redant.Command{
		Use:   "lava",
		Short: cliutil.UsageDesc("%s service", version.Project()),
		Long:  "Lava is a microservice integration framework",
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			fmt.Println("Usage: lava [command] [arguments]")
			fmt.Println("Available commands:")
			fmt.Println("  watch     Watch files for changes and run commands automatically")
			fmt.Println("  curl      Make HTTP requests to gRPC services")
			return nil
		},
		// 添加子命令
		Children: []*redant.Command{
			watchcmd.New(),
			curlcmd.New(),
		},
	}

	// 运行命令
	if err := app.Run(context.Background()); err != nil {
		os.Exit(1)
	}
}
