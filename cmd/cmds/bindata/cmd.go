package bindata

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"

	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/pkg/gutil"
	"github.com/pubgo/lug/pkg/shutil"
)

var Cmd = &cobra.Command{
	Use:     "bindata",
	Short:   "embed docs",
	Example: gutil.ExampleFmt(`lug bindata`),
	Run: func(cmd *cobra.Command, args []string) {
		defer xerror.RespExit()

		var shell = `go-bindata -fs -pkg docs -o docs/docs.go -prefix docs/ -ignore=docs\\.go docs/...`
		xerror.Panic(shutil.Bash(shell).Run())
		var code = gutil.CodeFormat(
			"package docs",
			`import "github.com/pubgo/lug/plugins/swagger"`,
			fmt.Sprintf("// build time: %s", time.Now().Format(consts.DefaultTimeFormat)),
			`func init() {swagger.Init(AssetNames, MustAsset)}`,
		)

		const path = "docs/init.go"
		_ = os.RemoveAll(path)
		xerror.Panic(ioutil.WriteFile(path, []byte(code), 0755))
	},
}
