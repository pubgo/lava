package main

import (
	"os"

	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/version"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/gateway/server"
)

func main() {
	var app = &cli.App{
		Name:    "lava-gateway",
		Version: version.Version,
		Flags:   flags.GetFlags(),
		Action: func(context *cli.Context) error {
			return server.Start()
		},
	}
	xerror.Exit(app.Run(os.Args))
}
