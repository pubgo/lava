package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/pubgo/lava/cmd/cmds/protoc"
	"github.com/pubgo/lava/cmd/cmds/restapi"
	"github.com/pubgo/lava/cmd/cmds/trace"
	"github.com/pubgo/lava/pkg/cli"
	"github.com/pubgo/lava/version"
)

func main() {
	cli.Execute(func(cmd *cobra.Command, flags *pflag.FlagSet) {
		cmd.Use = "lava"
		cmd.Version = version.Version

		cmd.AddCommand(trace.Cmd())
		cmd.AddCommand(protoc.Cmd())
		cmd.AddCommand(restapi.Cmd)
	})
}
