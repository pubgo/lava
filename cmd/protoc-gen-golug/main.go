package main

import (
	"log"

	"github.com/pubgo/golug/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xprotogen/gen"
)

func main() {
	m := gen.New("golug")
	m.Parameter(func(key, value string) {
		log.Println("params:", key, "=", value)
	})

	xerror.Exit(m.Init(func(fd *gen.FileDescriptor) {
		fd.Set("fdName", fd.GetName())

		j := fd.Jen
		j.PackageComment("// Code generated by protoc-gen-golug. DO NOT EDIT.")
		if !fd.GetOptions().GetDeprecated() {
			j.PackageComment("// source: " + fd.GetName())
		} else {
			j.PackageComment("// " + fd.GetName() + " is a deprecated file.")
		}

		j.Id(`
import "github.com/pubgo/golug/golug_data"
`)

		for _, ss := range fd.GetService() {
			var data = make(map[string]interface{})
			for _, m := range ss.GetMethod() {
				data[m.GetName()] = map[string]interface{}{
					"method":         m.P("{{.http_method}}"),
					"path":           m.P("{{.http_path}}"),
					"client_stream":  m.P("{{.cs}}") == "true",
					"server_streams": m.P("{{.ss}}") == "true",
				}
			}

			ss.Set("data", "`"+golug_util.Marshal(data)+"`")
			j.Id(ss.P(`func init() {golug_data.Add("{{.fdName}}.Register{{.srv}}Server",Register{{.srv}}Server)}`))
			j.Id(ss.P(`func init() {golug_data.Add("{{.fdName}}.{{.srv}}",{{.data}})}`))
		}
	}))
}
