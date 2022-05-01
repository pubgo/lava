package main

import (
	"flag"

	"github.com/pubgo/xerror"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/pubgo/lava/cmd/protoc-gen-resty/internal"
	"github.com/pubgo/lava/pkg/protoutil"
)

func main() {
	defer xerror.RespExit()

	var flags flag.FlagSet
	if protoutil.IsHelp() {
		flags.PrintDefaults()
		return
	}

	opts := &protogen.Options{ParamFunc: flags.Set}
	opts.Run(func(gen *protogen.Plugin) error {
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
