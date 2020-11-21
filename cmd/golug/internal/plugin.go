package internal

import (
	"github.com/pubgo/golug/golug_version"
	"github.com/spf13/cobra"
)

func NewPlugin() *cobra.Command {
	var cmd = &cobra.Command{Use: "plugin", Version: golug_version.Version}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		//	创建plugin相关的文件
	}

	return cmd
}
