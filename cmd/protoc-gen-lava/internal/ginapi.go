package internal

import (
	"net/http"
	"strings"

	"github.com/pubgo/lava/pkg/protoutil"
	"google.golang.org/protobuf/compiler/protogen"
)

func genGinRouter(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {
	if !genGin {
		return
	}

	g.P(`func Register`, service.GoName, `GinServer(r `, ginCall("IRouter"), `, server `, service.GoName, `Server) {`)
	g.P(xerrorCall("Assert"), `(r == nil || server == nil, "router or server is nil")`)
	for _, m := range service.Methods {
		// 过滤掉stream
		if m.Desc.IsStreamingClient() || m.Desc.IsStreamingServer() {
			continue
		}

		hr, err := protoutil.ExtractAPIOptions(m.Desc)
		if err != nil || hr == nil {
			hr = protoutil.DefaultAPIOptions(string(file.GoPackageName), service.GoName, m.GoName)
		}
		method, path := protoutil.ExtractHttpMethod(hr)
		method = strings.ToUpper(method)

		g.P(`r.Handle("`, method, `","`, path, `", func(ctx *`, ginCall("Context"), `) {`)
		g.P(`var req = new(`, g.QualifiedGoIdent(m.Input.GoIdent), `)`)
		if method == http.MethodGet {
			g.P(`xerror.Panic(`, bindingCall("MapFormByTag"), `(req, ctx.Request.URL.Query(), "json"))`)
		} else {
			g.P(`xerror.Panic(ctx.ShouldBindJSON(req))`)
		}
		g.P(`var resp,err=server.`, m.GoName, `(ctx,req)`)
		g.P(`xerror.Panic(err)`)
		g.P(`ctx.JSON(200,resp)`)
		g.P(`})`)
	}
	g.P(`}`)
}
