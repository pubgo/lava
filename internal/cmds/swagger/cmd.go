package swagger

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mattn/go-zglob/fastwalk"
	"github.com/pubgo/x/typex"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

var trim = strings.TrimSpace
var Cmd = &cobra.Command{
	Use:     "swagger",
	Aliases: typex.StrOf("v"),
	Short:   "gen swagger",
	Example: trim(`
lug swagger`),
	Run: func(cmd *cobra.Command, args []string) {
		defer xerror.RespExit()

		var path = "./docs/swagger"
		if len(args) > 0 {
			path = args[0]
		}

		var buf bytes.Buffer
		fmt.Fprintln(&buf, "package swagger")
		fmt.Fprintln(&buf, "var data = make(map[string]string)")
		xerror.Panic(fastwalk.FastWalk(path, func(path string, typ os.FileMode) error {
			if typ.IsDir() {
				return nil
			}

			var data = xerror.PanicBytes(ioutil.ReadFile(path))
			fmt.Fprintf(&buf, "func init() {data[`%s`]=`%s`}\n", path, hex.EncodeToString(data))

			return nil
		}))
		fmt.Fprintf(&buf, "func Data()map[string]string {return data}")

		var swagger = "./docs/swagger/swagger.go"
		_ = os.RemoveAll(swagger)
		xerror.Panic(ioutil.WriteFile(swagger, buf.Bytes(), 0755))
	},
}
