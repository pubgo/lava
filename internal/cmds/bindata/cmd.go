package bindata

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"

	"github.com/pubgo/lug/pkg/gutil"
	"github.com/pubgo/lug/pkg/shutil"
)

var trim = strings.TrimSpace
var Cmd = &cobra.Command{
	Use:   "bindata",
	Short: "embed docs",
	Example: trim(`
lug bindata`),
	Run: func(cmd *cobra.Command, args []string) {
		defer xerror.RespExit()

		for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
			if dir == "" {
				dir = "."
			}
			path := filepath.Join(dir, "go-bindata")
			if pathutil.IsExist(path) {
				xerror.Panic(shutil.Shell(path, `-fs -pkg docs -o docs/docs.go -prefix docs/ -ignore=docs\\.go docs/...`).Run())
			}
		}

		var code = gutil.CodeJoin(
			"package docs",
			`import (
				"github.com/pubgo/lug/plugins/swagger"
			)`,
			`func init() {
				swagger.Init(AssetNames, MustAsset)
			}`)

		const path = "docs/init.go"
		_ = os.RemoveAll(path)
		xerror.Panic(ioutil.WriteFile(path, []byte(code), 0755))
	},
}
