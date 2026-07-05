package gatewayserver

import (
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/merge"

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

// CombinedConfigLoader loads gateway_server and legacy grpc_server keys from YAML.
type CombinedConfigLoader struct {
	GatewayServer *Config `yaml:"gateway_server"`
	GrpcServer    *Config `yaml:"grpc_server"`
}

// ResolveConfig merges YAML loaders, preferring gateway_server and warning on legacy grpc_server.
func ResolveConfig(loader CombinedConfigLoader, logger log.Logger) *Config {
	switch {
	case loader.GatewayServer != nil && loader.GrpcServer != nil:
		if logger != nil {
			logger.Warn().Msg("both gateway_server and grpc_server are set; gateway_server wins — remove grpc_server from YAML")
		}
		return merge.Struct(defaultCfg(), loader.GatewayServer).Unwrap()
	case loader.GatewayServer != nil:
		return merge.Struct(defaultCfg(), loader.GatewayServer).Unwrap()
	case loader.GrpcServer != nil:
		if logger != nil {
			logger.Warn().Msg("yaml key grpc_server is deprecated; rename to gateway_server (see docs/legacy-removal.md)")
		}
		return merge.Struct(defaultCfg(), loader.GrpcServer).Unwrap()
	default:
		return defaultCfg()
	}
}

// LoadConfig reads gateway settings from the active funk config path, if any.
func LoadConfig(logger log.Logger) *Config {
	if config.GetConfigPath() == "" {
		return defaultCfg()
	}
	loaded := config.Load[CombinedConfigLoader]()
	return ResolveConfig(loaded.T, logger)
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
	// Default is true: register handlers on Mux only; native gRPC uses passthrough.
	// Set false for legacy dual registration on Mux and grpc.Server.
	GRPCPassthrough bool `yaml:"grpc_passthrough"`
}

func defaultCfg() *Config {
	return &Config{
		GRPCPassthrough: true,
	}
}
