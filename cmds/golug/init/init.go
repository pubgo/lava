package init

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/pubgo/golug/golug"
	"github.com/pubgo/golug/gutils"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

func NewInit() *cobra.Command {
	var cmd = &cobra.Command{Use: "init"}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		home := filepath.Join(xerror.PanicStr(homedir.Dir()), "."+golug.Project, "config")
		xerror.Panic(os.MkdirAll(home, 0755))

		fmt.Println("config home:", home)

		cfgPath := filepath.Join(home, "config.yaml")
		if gutils.PathExist(cfgPath) {
			return
		}

		xerror.Panic(ioutil.WriteFile(cfgPath, []byte(""), 0600))
	}

	return cmd
}
