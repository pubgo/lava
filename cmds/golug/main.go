package main

import (
	"github.com/pubgo/golug/cmds/golug/goimportdot"
	"github.com/pubgo/golug/cmds/golug/golist"
	"github.com/pubgo/golug/cmds/golug/gomod"
	"github.com/pubgo/golug/cmds/golug/initcmd"
	"github.com/pubgo/golug/cmds/golug/plugin"
	"github.com/pubgo/golug/version"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "golug", Version: version.Version, Short: "golug 项目管理命令"}
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { return xerror.Wrap(cmd.Help()) }
	rootCmd.AddCommand(initcmd.New())
	rootCmd.AddCommand(plugin.NewPlugin())
	rootCmd.AddCommand(gomod.GetCmd())
	rootCmd.AddCommand(goimportdot.GetCmd())
	rootCmd.AddCommand(golist.GetCmd())
	xerror.Exit(rootCmd.Execute())
}
