package gen_lava

import (
	"fmt"
	"github.com/pubgo/lava/cmd/protoc-gen-lava"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/pubgo/lava/pkg/protoutil"
)

// gen rest.http from protobuf
func genRestApiTest(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {
	var genPath = fmt.Sprintf("%s.%s.http", file.GoPackageName, service.GoName)

	var data []string
	for _, m := range service.Methods {
		hr, err := protoutil.ExtractAPIOptions(m.Desc)
		if err != nil || hr == nil {
			hr = protoutil.DefaultAPIOptions(string(file.GoPackageName), service.GoName, m.GoName)
		}
		method, path := protoutil.ExtractHttpMethod(hr)

		data = append(data, fmt.Sprintf("### %s.%s.%s\n", file.GoPackageName, service.GoName, m.GoName))

		if m.Comments.Leading.String() != "" {
			data = append(data, strings.TrimSpace(m.Comments.Leading.String())+"\n")
		}

		data = append(data, fmt.Sprintf("%s http://localhost:8080%s\n", method, path))
		data = append(data, fmt.Sprintf("Content-Type: application/json\n\n"))
	}
	xerror.Panic(pathutil.IsNotExistMkDir(main.testDir))
	xerror.Panic(ioutil.WriteFile(filepath.Join(main.testDir, genPath), []byte(strings.Join(data, "")), 0755))
}
