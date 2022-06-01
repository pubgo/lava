package internal

import (
	"fmt"
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
		method, url := protoutil.ExtractHttpMethod(hr)

		if m.Comments.Leading.String() != "" {
			data = append(data, strings.TrimSpace(m.Comments.Leading.String())+"\n")
		}
		data = append(data, fmt.Sprintf("### %s.%s.%s\n", file.GoPackageName, service.GoName, m.GoName))

		data = append(data, fmt.Sprintf("%s http://localhost:8080%s\n", method, url))
		data = append(data, fmt.Sprintf("Content-Type: application/json\n\n"))
	}
	xerror.Panic(pathutil.IsNotExistMkDir(testDir))
	xerror.Panic(ioutil.WriteFile(filepath.Join(testDir, genPath), []byte(strings.Join(data, "")), 0755))
}

//func genRestRouter(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {
//	if !genRest {
//		return
//	}
//
//	g.P(`func Dix`, service.GoName, `RestServer(app `, fiberCall("Router"), `, server `, service.GoName, `Server) {`)
//	g.P(xerrorCall("Assert"), `(app == nil || server == nil, "app or server is nil")`)
//	for _, m := range service.Methods {
//		if m.Desc.IsStreamingClient() || m.Desc.IsStreamingServer() {
//			continue
//		}
//
//		hr, err := protoutil.ExtractAPIOptions(m.Desc)
//		if err != nil || hr == nil {
//			hr = protoutil.DefaultAPIOptions(string(file.GoPackageName), service.GoName, m.GoName)
//		}
//		method, path := protoutil.ExtractHttpMethod(hr)
//		method = strings.ToUpper(method)
//
//		g.P(`app.Add("`, method, `","`, path, `", func(ctx *`, fiberCall("Ctx"), `) error {`)
//		g.P(`var req = new(`, g.QualifiedGoIdent(m.Input.GoIdent), `)`)
//		if method == http.MethodGet {
//			g.P(`data := make(map[string][]string)`)
//			g.P(`ctx.Context().QueryArgs().VisitAll(func(key []byte, val []byte) {`)
//			g.P(`	k := `, byteutilCall("ToStr"), `(key)`)
//			g.P(`	v := `, byteutilCall("ToStr"), `(val)`)
//			g.P(`	data[k] = append(data[k], v)`)
//			//g.P(`	if `, stringsCall("Contains"), `(v, ",") && `, bindingCall("EqualFieldType"), `(req, `, reflectCall("Slice"), `, k) {`)
//			//g.P(`		values := `, stringsCall("Split"), `(v, ",")`)
//			//g.P(`		for i := 0; i < len(values); i++ {`)
//			//g.P(`			data[k] = append(data[k], values[i])`)
//			//g.P(`		}`)
//			//g.P(`	} else {`)
//
//			//g.P(`	}`)
//			g.P(`})`)
//			g.P(`xerror.Panic(`, bindingCall("MapFormByTag"), `(req, data, "json"))`)
//		} else {
//			g.P(`xerror.Panic(ctx.BodyParser(req))`)
//		}
//		g.P(`var resp,err=server.`, m.GoName, `(ctx.UserContext(),req)`)
//		g.P(`xerror.Panic(err)`)
//		g.P(`return xerror.Wrap(ctx.JSON(resp))`)
//		g.P(`})`)
//	}
//	g.P(`}`)
//}
