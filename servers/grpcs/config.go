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
	// WebSocketPort enables the gateway WebSocket frontend on the given port.
	// When zero or unset, the WebSocket server is not started.
	WebSocketPort int64 `yaml:"websocket_port"`
	// WebSocketOriginPatterns sets allowed Origin patterns for the WS handshake.
	// See https://github.com/coder/websocket#origin-patterns
	WebSocketOriginPatterns []string `yaml:"websocket_origin_patterns"`
	// WebSocketInsecureSkipVerify disables WS origin checks (development only).
	WebSocketInsecureSkipVerify bool `yaml:"websocket_insecure_skip_verify"`
	// GRPCPassthrough enables native gRPC passthrough via Mux.UnknownServiceHandler.
	// When true, register services only on gateway.Mux; the outer grpc.Server
	// forwards all RPCs to the same backend used by HTTP/WebSocket frontends.
	GRPCPassthrough bool `yaml:"grpc_passthrough"`
}

func defaultCfg() *Config {
	return &Config{}
}
