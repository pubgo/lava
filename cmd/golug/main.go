package main

import (
	"github.com/pubgo/golug/cmd/golug/goimportdot"
	"github.com/pubgo/golug/cmd/golug/gomod"
	"github.com/pubgo/golug/cmd/golug/grpcurl"
	"github.com/pubgo/golug/cmd/golug/internal"
	"github.com/pubgo/golug/version"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "golug", Version: version.Version}
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { return xerror.Wrap(cmd.Help()) }
	rootCmd.AddCommand(internal.NewInit())
	rootCmd.AddCommand(internal.NewPlugin())
	rootCmd.AddCommand(grpcurl.GetCmd())
	rootCmd.AddCommand(gomod.GetCmd())
	rootCmd.AddCommand(goimportdot.GetCmd())
	xerror.Exit(rootCmd.Execute())
}
