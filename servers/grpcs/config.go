package grpcs

import (
	"github.com/pubgo/lava/v2/pkg/fiberbuilder"
	"github.com/pubgo/lava/v2/pkg/grpcbuilder"
)

const (
	defaultContentType = "application/grpc"
)

type GrpcServerConfigLoader struct {
	GrpcServer *Config `yaml:"grpc_server"`
}

type Config struct {
	Http              *fiberbuilder.Config `yaml:"http"`
	GrpcConfig        *grpcbuilder.Config  `yaml:"grpc"`
	EnablePrintRouter bool                 `yaml:"enable_print_router"`
}

func defaultCfg() *Config {
	return &Config{}
}
