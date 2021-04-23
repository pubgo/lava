package golist

import (
	"fmt"
	"strings"

	"github.com/pubgo/x/shutil"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

func GetCmd() *cobra.Command {
	var cmd = &cobra.Command{Use: "golist"}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		var filter = map[string]bool{
			"github.com/pubgo/lugin":         true,
			"github.com/pubgo/lugin/scripts": true,
			"github.com/pubgo/lugin/version": true,
			"github.com/pubgo/lugin/example": true,
		}

		for _, v := range strings.Split(xerror.PanicStr(shutil.GoList()), "\n") {
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

			if strings.Contains(v, "github.com/pubgo/lugin/") {
				fmt.Println(v)
			}
		}
	}
	return cmd
}
