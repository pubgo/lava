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
	// ZrpcURL enables the gateway NATS/zrpc frontend when non-empty (requires ZrpcQueue).
	// Example: nats://127.0.0.1:4222
	ZrpcURL string `yaml:"zrpc_url"`
	// ZrpcQueue is the NATS queue group for all gateway zrpc method bindings.
	ZrpcQueue string `yaml:"zrpc_queue"`
	// ZrpcSubjectPrefix is prepended to each gRPC full method for the NATS subject.
	// Default when empty: "svc." — "/pkg.v1.Service/Method" → "svc.pkg.v1.Service/Method".
	ZrpcSubjectPrefix string `yaml:"zrpc_subject_prefix"`
}

func defaultCfg() *Config {
	return &Config{}
}
