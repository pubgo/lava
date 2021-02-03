package plugin

import (
	"github.com/spf13/cobra"
)

func NewPlugin() *cobra.Command {
	var cmd = &cobra.Command{Use: "plugin"}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		//	创建plugin相关的文件
	}

	return cmd
}
