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
	Grpc       *grpc_builder.Config `yaml:"grpc_config"`
}

func defaultCfg() Config {
	return Config{
		Grpc:    grpc_builder.GetDefaultCfg(),
		BaseUrl: version.Project(),
	}
}
