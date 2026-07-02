package gatewayserver

import (
	"github.com/pubgo/lava/v2/pkg/fiberbuilder"
	"github.com/pubgo/lava/v2/pkg/grpcbuilder"
)

const (
	defaultContentType = "application/grpc"
)

// ConfigLoader loads gateway server settings from gateway_server YAML key.
type ConfigLoader struct {
	GatewayServer *Config `yaml:"gateway_server"`
}

// GrpcServerConfigLoader is a legacy alias for grpc_server YAML key.
type GrpcServerConfigLoader struct {
	GrpcServer *Config `yaml:"grpc_server"`
}

// Config holds HTTP/REST, WebSocket, and native gRPC gateway server settings.
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
	// Default is false for backward compatibility with existing deployments that
	// register services on both Mux and grpc.Server.
	GRPCPassthrough bool `yaml:"grpc_passthrough"`
}

func defaultCfg() *Config {
	return &Config{}
}
