package clix

import (
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
)

func Command(cb func(cmd *cobra.Command, flags *pflag.FlagSet)) *cobra.Command {
	var cmd = &cobra.Command{}
	cb(cmd, cmd.PersistentFlags())
	return cmd
}

func Execute(cb func(cmd *cobra.Command, flags *pflag.FlagSet)) {
	defer xerror.RespExit()

	var cmd = &cobra.Command{}
	cb(cmd, cmd.Flags())
	xerror.Panic(cmd.Execute())
}

func ExampleFmt(data ...string) string {
	var str = ""
	for i := range data {
		str += "  " + data[i] + "\n"
	}
	return "  " + strings.TrimSpace(str)
}