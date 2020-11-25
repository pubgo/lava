package main

import (
	"github.com/pubgo/golug/cmd/golug/grpcurl"
	"github.com/pubgo/golug/cmd/golug/internal"
	"github.com/pubgo/golug/golug_version"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "golug", Version: golug_version.Version}
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { return xerror.Wrap(cmd.Help()) }
	rootCmd.AddCommand(internal.NewInit())
	rootCmd.AddCommand(internal.NewPlugin())
	rootCmd.AddCommand(grpcurl.GetCmd())
	xerror.Exit(rootCmd.Execute())
}
