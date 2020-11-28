package golug

import (
	"github.com/spf13/cobra"
)

func NewInit() *cobra.Command {
	var cmd = &cobra.Command{Use: "init"}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		//	创建plugin相关的文件
	}

	return cmd
}
