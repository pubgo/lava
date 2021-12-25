package internal

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/pubgo/xerror"
	"google.golang.org/protobuf/compiler/protogen"
	gp "google.golang.org/protobuf/proto"

	"github.com/pubgo/lava/errors"
	"github.com/pubgo/lava/pkg/protoutil"
)

var (
	contextCall  = protoutil.Import("context")
	reflectCall  = protoutil.Import("reflect")
	stringsCall  = protoutil.Import("strings")
	sqlCall      = protoutil.Import("database/sql")
	sqlxCall     = protoutil.Import("gorm.io/gorm")
	grpcCall     = protoutil.Import("google.golang.org/grpc")
	codesCall    = protoutil.Import("google.golang.org/grpc/codes")
	statusCall   = protoutil.Import("google.golang.org/grpc/status")
	grpccCall    = protoutil.Import("github.com/pubgo/lava/clients/grpcc")
	xerrorCall   = protoutil.Import("github.com/pubgo/xerror")
	errorsCall   = protoutil.Import("github.com/pubgo/lava/errors")
	fiberCall    = protoutil.Import("github.com/pubgo/lava/builder/fiber")
	ginCall      = protoutil.Import("github.com/gin-gonic/gin")
	bindingCall  = protoutil.Import("github.com/pubgo/lava/pkg/binding")
	byteutilCall = protoutil.Import("github.com/pubgo/x/byteutil")
)

// GenerateFile generates a .lava.pb.go file containing service definitions.
func GenerateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Enums) == 0 {
		return nil
	}

	filename := file.GeneratedFilenamePrefix + ".errors.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-errors. DO NOT EDIT.")
	g.P("// versions:")
	g.P("// - protoc-gen-errors ", version)
	g.P("// - protoc            ", protocVersion(gen))
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()

	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the grpc package it is being compiled against.")
	g.P("// Requires gRPC-Go v1.32.0 or later.")
	g.P("const _ = ", grpcCall("SupportPackageIsVersion7"))
	g.P()

	generateFileContent(gen, file, g)
	return g
}

// generateFileContent generates the service definitions, excluding the package statement.
func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	index := 0
	for _, enum := range file.Enums {
		if genError(gen, file, g, enum) {
			index++
		}
	}

	if index == 0 {
		g.Skip()
		return
	}

	g.QualifiedGoIdent(errorsCall(""))
}

func genError(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, enum *protogen.Enum) (ret bool) {
	defer xerror.RespRaise(func(err xerror.XErr) error { ret = false; return err })

	var content = `
func Is{{.camelValue}}(err error) bool {
	e := errors.FromError(err)
	return e.Reason == {{.name}}_{{.value}}.String() && e.Code == {{.code}} 
}

func Error{{.camelValue}}(format string, args ...interface{}) *errors.Error {
	 return errors.New({{.name}}_{{.value}}.String(), {{.code}}, format, args...)
}
`
	tmpl, err := template.New("errors").Parse(content)
	xerror.Panic(err)

	var isOk bool
	for _, v := range enum.Values {
		var opts = v.Desc.Options()
		if !gp.HasExtension(opts, errors.E_Code) {
			continue
		}

		isOk = true
		code, ok := gp.GetExtension(opts, errors.E_Code).(int32)
		xerror.Assert(!ok, "errors.code type error")

		// 业务code大于1000
		xerror.Assert(code < 1000, "code must be greater than 1000, now(%d)", code)

		xerror.Panic(tmpl.Execute(g, map[string]interface{}{
			"name":       string(enum.Desc.Name()),
			"value":      string(v.Desc.Name()),
			"camelValue": case2Camel(string(v.Desc.Name())),
			"code":       code,
		}))
	}

	return isOk
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

func case2Camel(name string) string {
	if !strings.Contains(name, "_") {
		return strings.Title(strings.ToLower(name))
	}
	name = strings.ToLower(name)
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Title(name)
	return strings.Replace(name, " ", "", -1)
}