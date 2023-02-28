package grpcs

import (
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/internal/grpc_builder"
)

const (
	Name               = "service"
	defaultContentType = "application/grpc"
)

type Config struct {
	PathPrefix string               `yaml:"path-prefix"`
	Grpc       *grpc_builder.Config `yaml:"grpc-server"`
	PrintRoute bool                 `yaml:"print-route"`
}

func defaultCfg() Config {
	return Config{
		Grpc:       grpc_builder.GetDefaultCfg(),
		PrintRoute: true,
		PathPrefix: version.Project(),
	}
}
