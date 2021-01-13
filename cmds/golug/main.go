package main

import (
	"github.com/pubgo/golug/cmds/golug/goimportdot"
	"github.com/pubgo/golug/cmds/golug/golist"
	"github.com/pubgo/golug/cmds/golug/golug"
	"github.com/pubgo/golug/cmds/golug/gomod"
	"github.com/pubgo/golug/cmds/golug/grpcall"
	"github.com/pubgo/golug/cmds/golug/grpcurl"
	"github.com/pubgo/golug/version"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "golug", Version: version.Version}
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { return xerror.Wrap(cmd.Help()) }
	rootCmd.AddCommand(golug.NewInit())
	rootCmd.AddCommand(golug.NewPlugin())
	rootCmd.AddCommand(grpcurl.GetCmd())
	rootCmd.AddCommand(gomod.GetCmd())
	rootCmd.AddCommand(goimportdot.GetCmd())
	rootCmd.AddCommand(grpcall.GetCmd())
	rootCmd.AddCommand(golist.GetCmd())
	xerror.Exit(rootCmd.Execute())
}
