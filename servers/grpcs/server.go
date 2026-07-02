package grpcs

import (
	"github.com/pubgo/lava/v2/core/supervisor"
	"github.com/pubgo/lava/v2/servers/gatewayserver"
)

// Config is an alias of gatewayserver.Config (legacy grpc_server YAML key).
type Config = gatewayserver.Config

// Params is an alias of gatewayserver.Params.
type Params = gatewayserver.Params

// GrpcServerConfigLoader loads grpc_server YAML configuration (legacy).
type GrpcServerConfigLoader = gatewayserver.GrpcServerConfigLoader

// New creates a grpc-server supervisor service (legacy name for gatewayserver).
//
// Deprecated: use gatewayserver.New for new code.
func New(params Params) supervisor.Service {
	return gatewayserver.NewWithName(params, "grpc-server")
}
