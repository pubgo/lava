package main

import (
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/pubgo/lava/v2/internal/zrpcgen"
)

func main() {
	protogen.Options{}.Run(zrpcgen.Generate)
}
