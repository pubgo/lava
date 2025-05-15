package grpcs

import (
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/pkg/grpcbuilder"
)

const (
	defaultContentType = "application/grpc"
)

type GrpcServerConfigLoader struct {
	GrpcServer *Config `yaml:"grpc_server"`
}

type Config struct {
	EnablePrintRoutes bool                `yaml:"enable_print_routes"`
	BaseUrl           string              `yaml:"base_url"`
	GrpcConfig        *grpcbuilder.Config `yaml:"grpc_config"`
	EnableCors        bool                `yaml:"enable_cors"`
	EnablePingPong    bool                `yaml:"enable_ping_pong"`

	// unix seconds
	PingPongTime int32 `yaml:"ping_pong_time"`
	GrpcPort     *int  `yaml:"grpc_port"`
	HttpPort     *int  `yaml:"http_port"`

	WsReadLimit *int `yaml:"ws_read_limit"`
}

func defaultCfg() *Config {
	return &Config{
		EnablePrintRoutes: true,
		BaseUrl:           version.Project(),
		GrpcConfig:        grpcbuilder.GetDefaultCfg(),
	}
}
