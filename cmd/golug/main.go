package main

import (
	"github.com/pubgo/golug/cmd/golug/entc"
	"github.com/pubgo/golug/cmd/golug/goimportdot"
	"github.com/pubgo/golug/cmd/golug/golug"
	"github.com/pubgo/golug/cmd/golug/gomod"
	"github.com/pubgo/golug/cmd/golug/grpcall"
	"github.com/pubgo/golug/cmd/golug/grpcurl"
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
	rootCmd.AddCommand(entc.GetCmd())
	xerror.Exit(rootCmd.Execute())
}
