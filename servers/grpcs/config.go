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
	BaseUrl        string               `yaml:"base_url"`
	GrpcConfig     *grpc_builder.Config `yaml:"grpc_config"`
	EnableCors     bool                 `yaml:"enable_cors"`
	EnablePingPong bool                 `yaml:"enable_ping_pong"`

	// unix seconds
	PingPongTime int32 `yaml:"ping_pong_time"`
	GrpcPort     *int  `yaml:"grpc_port"`
	HttpPort     *int  `yaml:"http_port"`

	WsReadLimit *int `yaml:"ws_read_limit"`
}

func defaultCfg() *Config {
	return &Config{
		BaseUrl:    version.Project(),
		GrpcConfig: grpc_builder.GetDefaultCfg(),
	}
}
