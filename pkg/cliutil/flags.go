package cliutil

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Flags(cmd *cobra.Command, cb func(flags *pflag.FlagSet)) {
	cb(cmd.Flags())
}
