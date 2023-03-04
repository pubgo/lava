package grpcs

import (
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/internal/grpc_builder"
)

const (
	defaultContentType = "application/grpc"
)

type Config struct {
	PrintRoute bool                 `yaml:"print_route"`
	BaseUrl    string               `yaml:"base_url"`
	GrpcConfig *grpc_builder.Config `yaml:"grpc_config"`
}

func defaultCfg() Config {
	return Config{
		PrintRoute: true,
		BaseUrl:    version.Project(),
		GrpcConfig: grpc_builder.GetDefaultCfg(),
	}
}
