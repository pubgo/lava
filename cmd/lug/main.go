package main

import (
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"

	"github.com/pubgo/lug/cmd/cmds/protoc"
	"github.com/pubgo/lug/cmd/cmds/restapi"
	"github.com/pubgo/lug/cmd/cmds/trace"
	"github.com/pubgo/lug/version"
)

func main() {
	defer xerror.RespExit()

	var rootCmd = &cobra.Command{Use: "lug", Version: version.Version}
	rootCmd.AddCommand(trace.Cmd())
	rootCmd.AddCommand(protoc.Cmd())
	rootCmd.AddCommand(restapi.Cmd)
	xerror.Panic(rootCmd.Execute())
}
