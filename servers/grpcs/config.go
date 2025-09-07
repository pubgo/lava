package grpcs

import (
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/running"

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
	HttpPort          *int                 `yaml:"http_port"`
	GrpcConfig        *grpcbuilder.Config  `yaml:"grpc"`
	GrpcPort          *int                 `yaml:"grpc_port"`
	EnablePrintRouter bool                 `yaml:"enable_print_router"`
	BaseUrl           string               `yaml:"base_url"`
}

func defaultCfg() *Config {
	return &Config{
		GrpcConfig: grpcbuilder.GetDefaultCfg(),
		GrpcPort:   generic.Ptr(running.GrpcPort),
	}
}
