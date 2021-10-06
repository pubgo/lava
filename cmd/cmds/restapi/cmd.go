package restapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"

	"github.com/pubgo/lug/pkg/gutil"
	"github.com/pubgo/lug/xgen"
)

var Cmd = &cobra.Command{
	Use:     "rest.http",
	Short:   "gen rest.http from protobuf",
	Example: gutil.ExampleFmt(`lug rest.http`),
	Run: func(cmd *cobra.Command, args []string) {
		defer xerror.RespExit()

		for _, val := range xgen.List() {
			if val == nil {
				continue
			}

			handlers, ok := val.([]xgen.GrpcRestHandler)
			if !ok {
				continue
			}

			var data []string
			var name = ""
			for _, handler := range handlers {
				name = handler.Service
				data = append(data, fmt.Sprintf("### %s.%s\n", handler.Service, handler.Name))
				data = append(data, fmt.Sprintf("%s http://localhost:8080%s\n", handler.Method, handler.Path))
				data = append(data, fmt.Sprintf("Content-Type: application/json\n\n"))
				if handler.Method != http.MethodGet {
					var params = make(map[string]string)
					var tt = reflect.TypeOf(handler.Input).Elem()
					for i := tt.NumField() - 1; i >= 0; i-- {
						var f = tt.Field(i)
						var tag = f.Tag.Get("json")
						if tag != "" {
							params[strings.Split(tag, ",")[0]] = ""
						}
					}
					data = append(data, fmt.Sprintf("%s\n\n", xerror.PanicBytes(json.MarshalIndent(params, "", " "))))
				}
			}
			name = fmt.Sprintf("tests/http/%s.http", strings.ToLower(name))
			xerror.Panic(pathutil.IsNotExistMkDir(filepath.Dir(name)))
			xerror.Panic(ioutil.WriteFile(name, []byte(strings.Join(data, "")), 0755))
		}
	},
}