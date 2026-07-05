package grpcs

import (
	"sync"

	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/supervisor"
	"github.com/pubgo/lava/v2/servers/gatewayserver"
)

var deprecateOnce sync.Once

// Config is an alias of gatewayserver.Config (legacy grpc_server YAML key).
type Config = gatewayserver.Config

// Params is an alias of gatewayserver.Params.
type Params = gatewayserver.Params

// GrpcServerConfigLoader loads grpc_server YAML configuration (legacy).
type GrpcServerConfigLoader = gatewayserver.GrpcServerConfigLoader

// CombinedConfigLoader is an alias of gatewayserver.CombinedConfigLoader.
type CombinedConfigLoader = gatewayserver.CombinedConfigLoader

// New creates a grpc-server supervisor service (legacy name for gatewayserver).
//
// Deprecated: use gatewayserver.New for new code.
func New(params Params) supervisor.Service {
	deprecateOnce.Do(func() {
		log.Warn().Msg("servers/grpcs is deprecated; import servers/gatewayserver instead (removed in v3)")
	})
	return gatewayserver.NewWithName(params, "grpc-server")
}
