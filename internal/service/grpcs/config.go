package grpcs

import (
	"github.com/pubgo/lava/internal/pkg/fiber_builder"
	"github.com/pubgo/lava/internal/pkg/grpc_builder"
)

const (
	Name               = "service"
	defaultContentType = "application/grpc"
)

type GrpcWebCfg struct {
}

type Cfg struct {
	Grpc       *grpc_builder.Config `yaml:"grpc-server"`
	Api        *fiber_builder.Cfg   `yaml:"http-server"`
	GrpcWeb    *GrpcWebCfg          `yaml:"grpc-web"`
	PrintRoute bool                 `yaml:"print-route"`
}
