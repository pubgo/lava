package main

import (
	"flag"

	"github.com/pubgo/lava/component/cloudjobs/protoc-gen-cloud-job/internal"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	flag.Parse()

	protogen.Options{ParamFunc: flag.CommandLine.Set}.
		Run(func(gp *protogen.Plugin) error {
			gp.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

			for _, name := range gp.Request.FileToGenerate {
				internal.GenerateFile(gp, gp.FilesByPath[name])
			}

			return nil
		})
}
