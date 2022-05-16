package internal

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/pubgo/lava/pkg/protoutil"
)

var (
	contextCall  = protoutil.Import("context")
	serviceCall  = protoutil.Import("github.com/pubgo/lava/service")
	reflectCall  = protoutil.Import("reflect")
	stringsCall  = protoutil.Import("strings")
	sqlCall      = protoutil.Import("database/sql")
	sqlxCall     = protoutil.Import("gorm.io/gorm")
	dixCall      = protoutil.Import("github.com/pubgo/dix")
	grpcCall     = protoutil.Import("google.golang.org/grpc")
	codesCall    = protoutil.Import("google.golang.org/grpc/codes")
	statusCall   = protoutil.Import("google.golang.org/grpc/status")
	xerrorCall   = protoutil.Import("github.com/pubgo/xerror")
	xgenCall     = protoutil.Import("github.com/pubgo/lava/xgen")
	fiberCall    = protoutil.Import("github.com/pubgo/lava/builder/fiber")
	ginCall      = protoutil.Import("github.com/gin-gonic/gin")
	bindingCall  = protoutil.Import("github.com/pubgo/lava/pkg/binding")
	byteutilCall = protoutil.Import("github.com/pubgo/x/byteutil")
	runtimeCall  = protoutil.Import("github.com/grpc-ecosystem/grpc-gateway/v2/runtime")
	injectCall   = protoutil.Import("github.com/pubgo/lava/inject")
	grpccCall    = protoutil.Import("github.com/pubgo/lava/clients/grpcc")
	grpccCfgCall = protoutil.Import("github.com/pubgo/lava/clients/grpcc/grpcc_config")
	configCall   = protoutil.Import("github.com/pubgo/lava/config")
)

// GenerateFile generates a .lava.pb.go file containing service definitions.
func GenerateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 {
		return nil
	}

	if !enableLava {
		return nil
	}

	filename := file.GeneratedFilenamePrefix + ".lava.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-lava. DO NOT EDIT.")
	g.P("// versions:")
	g.P("// - protoc-gen-lava ", version)
	g.P("// - protoc         ", protocVersion(gen))
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()

	generateFileContent(gen, file, g)
	return g
}

// generateFileContent generates the service definitions, excluding the package statement.
func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	if len(file.Services) == 0 {
		return
	}

	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the grpc package it is being compiled against.")
	g.P("// Requires gRPC-Go v1.32.0 or later.")
	g.P("const _ = ", grpcCall("SupportPackageIsVersion7"))
	g.P()
	for _, service := range file.Services {
		genClient(gen, file, g, service)
		genRpcInfo(gen, file, g, service)
		//genRestApiTest(gen, file, g, service)
		//genRestRouter(gen, file, g, service)
		//genGinRouter(gen, file, g, service)
		//genSql(gen, file, g, service)
	}
}

func genClient(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {
	g.QualifiedGoIdent(injectCall(""))
	g.QualifiedGoIdent(grpccCall(""))
	g.QualifiedGoIdent(grpccCfgCall(""))
	g.QualifiedGoIdent(configCall(""))
	g.QualifiedGoIdent(xerrorCall(""))
	g.P(protoutil.Template(`
		func init() {
	xerror.RespExit()
	var cfgMap = make(map[string]*grpcc_config.Cfg)
	xerror.Panic(config.Decode(grpcc_config.Name, cfgMap))
	for name := range cfgMap {
		var cfg = cfgMap[name]
		var addr = name
		inject.RegName(cfg.Alias, func() {{name}}Client {
			return New{{name}}Client(grpcc.NewClient(addr))
		})
	}
}`, protoutil.Context{"name": service.GoName}))
	g.P()
}

func genRpcInfo(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {
	g.P("func init(){")
	g.P("var mthList []", xgenCall("GrpcRestHandler"))

	var isGw bool
	for _, m := range service.Methods {
		g.P("mthList = append(mthList, ", xgenCall("GrpcRestHandler"), "{")
		g.P("Input:        &", g.QualifiedGoIdent(m.Input.GoIdent), "{},")
		g.P("Output:        &", g.QualifiedGoIdent(m.Output.GoIdent), "{},")
		g.P(fmt.Sprintf(`Service:"%s",`, service.Desc.FullName()))
		g.P(fmt.Sprintf(`Name:"%s",`, m.Desc.Name()))

		var defaultUrl bool
		hr, err := protoutil.ExtractAPIOptions(m.Desc)
		if err == nil && hr != nil {
			defaultUrl = true
			isGw = true

			var replacer = strings.NewReplacer(".", "/", "-", "/")
			hr = protoutil.DefaultAPIOptions(replacer.Replace(string(file.Desc.Package())), service.GoName, m.GoName)
		}
		method, path := protoutil.ExtractHttpMethod(hr)
		g.P(fmt.Sprintf(`Method:"%s",`, method))
		g.P(fmt.Sprintf(`Path:"%s",`, path))
		g.P(fmt.Sprintf(`DefaultUrl:%v,`, defaultUrl))
		g.P("ClientStream:", m.Desc.IsStreamingClient(), ",")
		g.P("ServerStream:", m.Desc.IsStreamingServer(), ",")
		g.P("})")
		g.P()
	}
	// grpc
	g.P(xgenCall("Add"), "(Register", service.GoName, "Server, mthList)")
	g.P("}")
	g.P()

	if enableLava {
		if isGw {
			g.QualifiedGoIdent(grpcCall(""))
			g.QualifiedGoIdent(runtimeCall(""))
			g.QualifiedGoIdent(contextCall(""))
			g.QualifiedGoIdent(serviceCall(""))
		}
		g.P(protoutil.Template(`
func Register{{name}}(srv service.Service, impl {{name}}Server) {
	srv.RegService(service.Desc{
		Handler:     impl,
		ServiceDesc: {{name}}_ServiceDesc,
	})
	{% if isGw %}
	srv.RegGateway(func(ctx context.Context, mux *runtime.ServeMux, cc grpc.ClientConnInterface) error {
		return Register{{name}}HandlerClient(ctx, mux, New{{name}}Client(cc))
	})
    {% endif %}
}
`, protoutil.Context{"name": service.GoName, "isGw": isGw}))
		g.P()
	}
}

func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}
