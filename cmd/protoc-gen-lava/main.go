package main

import (
	"github.com/pubgo/xerror"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	genLava "github.com/pubgo/lava/cmd/protoc-gen-lava/internal"
)

func main() {
	defer xerror.RecoverAndExit()
	protogen.Options{ParamFunc: genLava.Flags.Set}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			genLava.GenerateFile(gen, f)
			genLava.GenerateTag(gen, f)
		}
		return nil
	})
}
