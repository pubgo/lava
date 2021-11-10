package main

import (
	"flag"

	"github.com/pubgo/xerror"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "v0.1.0"

var genRest bool
var genGin bool

func main() {
	defer xerror.RespExit()

	var flags flag.FlagSet
	flags.BoolVar(&genRest, "rest", false, "generate rest api")
	flags.BoolVar(&genGin, "gin", false, "generate gin api")
	protogen.Options{ParamFunc: flags.Set}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			generateFile(gen, f)
			generateTag(gen, f)
		}
		return nil
	})
}