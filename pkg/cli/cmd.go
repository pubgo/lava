package cli

import (
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(cb func(cmd *cobra.Command, flags *pflag.FlagSet)) *cobra.Command {
	var cmd = &cobra.Command{}
	cb(cmd, cmd.Flags())
	return cmd
}

func Execute(cb func(cmd *cobra.Command, flags *pflag.FlagSet)) {
	defer xerror.RespExit()

	var cmd = &cobra.Command{}
	cb(cmd, cmd.Flags())
	xerror.Panic(cmd.Execute())
}
