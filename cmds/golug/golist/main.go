package golist

import (
	"fmt"
	"strings"

	"github.com/pubgo/golug/pkg/golug_sh"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

func GetCmd() *cobra.Command {
	var cmd = &cobra.Command{Use: "golist"}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		var filter = map[string]bool{
			"github.com/pubgo/golugin":         true,
			"github.com/pubgo/golugin/scripts": true,
			"github.com/pubgo/golugin/version": true,
			"github.com/pubgo/golugin/example": true,
		}

		for _, v := range strings.Split(xerror.PanicStr(golug_sh.GoList()), "\n") {
			if strings.Contains(v, "internal") {
				continue
			}

			if strings.Contains(v, "example") {
				continue
			}

			if strings.Contains(v, "util") {
				continue
			}

			if strings.Contains(v, "/pkg/") {
				continue
			}

			if strings.Contains(v, "/cmd") {
				continue
			}

			if strings.Contains(v, "vendor") {
				continue
			}

			if filter[v] {
				continue
			}

			if v == "" {
				continue
			}

			if strings.Contains(v, "github.com/pubgo/golugin/") {
				fmt.Println(v)
			}
		}
	}
	return cmd
}
