package grpcs

import (
	"github.com/pubgo/funk/version"

	"github.com/pubgo/lava/pkg/grpc_builder"
)

const (
	defaultContentType = "application/grpc"
)

type GrpcServerConfigLoader struct {
	GrpcServer *Config `yaml:"grpc_server"`
}

type Config struct {
	EnablePrintRoutes bool                 `yaml:"enable_print_routes"`
	BaseUrl           string               `yaml:"base_url"`
	GrpcConfig        *grpc_builder.Config `yaml:"grpc_config"`
}

func defaultCfg() *Config {
	return &Config{
		EnablePrintRoutes: true,
		BaseUrl:           version.Project(),
		GrpcConfig:        grpc_builder.GetDefaultCfg(),
	}
}
