package main

import (
	"github.com/pubgo/xerror"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/pubgo/lava/cmd/protoc-gen-errors/internal"
)

func main() {
	defer xerror.RespExit()
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			internal.GenerateFile(gen, f)
		}
		return nil
	})
}
