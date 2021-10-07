package main

import (
	"flag"

	"github.com/pubgo/xerror"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "v0.1.0"

var genRest bool

func main() {
	defer xerror.RespExit()

	var flags flag.FlagSet
	flags.BoolVar(&genRest, "rest", false, "generate rest api")
	protogen.Options{ParamFunc: flags.Set}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			generateFile(gen, f)
		}
		return nil
	})
}
