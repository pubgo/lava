package main

import (
	"github.com/pubgo/golug/cmd/golug/internal"
	"github.com/pubgo/golug/golug_version"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "golug", Version: golug_version.Version}

	rootCmd.AddCommand(internal.NewPlugin())
	xerror.Exit(rootCmd.Execute())
}
