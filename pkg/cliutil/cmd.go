package cliutil

import "github.com/spf13/cobra"

func Cmd(cb func(cmd *cobra.Command)) *cobra.Command {
	var cmd = &cobra.Command{}
	cb(cmd)
	return cmd
}
