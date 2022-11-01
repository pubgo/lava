package main

import (
	"fmt"

	"github.com/pubgo/funk/recovery"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		defer recovery.Exit()

		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			fmt.Println(f)
		}
		return nil
	})
}

//func run() error {
//	req, err := plugin.ReadRequest(os.Stdin)
//	if err != nil {
//		return err
//	}
//
//	var plugin string
//
//	opts := protogen.Options{
//		ParamFunc: func(name, value string) error {
//			switch name {
//			case "plugin":
//				plugin = value
//			}
//			return nil // Ignore unknown params.
//		},
//	}
//
//	gen, err := opts.New(req)
//	if err != nil {
//		return err
//	}
//
//	if plugin == "" {
//		s := strings.TrimPrefix(filepath.Base(os.Args[0]), "protoc-gen-")
//		return fmt.Errorf("no protoc plugin specified; use 'protoc --%s_out=plugin=$PLUGIN:...'", s)
//	}
//
//	if os.Getenv("PROTO_PATCH_DEBUG_LOGGING") == "" {
//		log.SetOutput(ioutil.Discard)
//	}
//
//	// Strip our custom param(s).
//	plugin.StripParam(gen.Request, "plugin")
//
//	// Run the specified plugin and unmarshal the CodeGeneratorResponse.
//	res, err := plugin.RunPlugin(plugin, gen.Request, nil)
//	if err != nil {
//		return err
//	}
//
//	// Initialize a Patcher and scan source proto files.
//	patcher, err := patch.NewPatcher(gen)
//	if err != nil {
//		return err
//	}
//
//	// Patch the CodeGeneratorResponse.
//	err = patcher.Patch(res)
//	if err != nil {
//		return err
//	}
//
//	supportedFeatures := uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
//	res.SupportedFeatures = &supportedFeatures
//
//	// Write the patched CodeGeneratorResponse to stdout.
//	return plugin.WriteResponse(os.Stdout, res)
//}
