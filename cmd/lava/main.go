package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/pubgo/lava/cmd/cmds/protoc"
	"github.com/pubgo/lava/cmd/cmds/restapi"
	"github.com/pubgo/lava/cmd/cmds/swagger"
	"github.com/pubgo/lava/cmd/cmds/trace"
	"github.com/pubgo/lava/pkg/clix"
	"github.com/pubgo/lava/version"
)

func main() {
	clix.Execute(func(cmd *cobra.Command, flags *pflag.FlagSet) {
		cmd.Use = "lava"
		cmd.Version = version.Version

		cmd.AddCommand(trace.Cmd())
		cmd.AddCommand(protoc.Cmd())
		cmd.AddCommand(swagger.Cmd)
		cmd.AddCommand(restapi.Cmd)
	})
}
