package main

import (
	"github.com/pubgo/lug/cmd/lug/goimportdot"
	"github.com/pubgo/lug/cmd/lug/golist"
	"github.com/pubgo/lug/cmd/lug/gomod"
	"github.com/pubgo/lug/cmd/lug/initcmd"
	"github.com/pubgo/lug/cmd/lug/plugin"
	"github.com/pubgo/lug/version"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "lug", Version: version.Version, Short: "lug框架管理工具"}
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { return xerror.Wrap(cmd.Help()) }
	rootCmd.AddCommand(initcmd.New())
	rootCmd.AddCommand(plugin.NewPlugin())
	rootCmd.AddCommand(gomod.GetCmd())
	rootCmd.AddCommand(goimportdot.GetCmd())
	rootCmd.AddCommand(golist.GetCmd())
	xerror.Exit(rootCmd.Execute())
}
