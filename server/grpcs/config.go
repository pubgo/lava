package grpcs

import (
	"github.com/pubgo/lava/pkg/grpc_builder"
)

const (
	Name               = "service"
	defaultContentType = "application/grpc"
)

type GrpcWebCfg struct {
}

type Cfg struct {
	Grpc       *grpc_builder.Config `yaml:"grpc-server"`
	GrpcWeb    *GrpcWebCfg          `yaml:"grpc-web"`
	PrintRoute bool                 `yaml:"print-route"`
	BasePrefix string               `json:"base-prefix"`
}