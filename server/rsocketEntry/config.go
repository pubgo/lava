package rsocketEntry

import (
	"github.com/rsocket/rsocket-go/core/transport"
	"time"

	"github.com/pubgo/lava/server/grpcEntry/grpcs"
)

const Name = "rsocket_entry"

const (
	// DefaultMaxMsgSize define maximum message size that server can send or receive.
	// Default value is 4MB.
	DefaultMaxMsgSize = 1024 * 1024 * 4

	DefaultSleepAfterDeRegister = time.Second * 2

	// DefaultRegisterTTL The register expiry time
	DefaultRegisterTTL = time.Minute

	// DefaultRegisterInterval The interval on which to register
	DefaultRegisterInterval = time.Second * 30

	defaultContentType = "application/grpc"

	DefaultSleepAfterDeregister = time.Second * 2
)

type Cfg struct {
	GrpcWeb              bool          `json:"grpc_web"`
	Grpc                 *grpcs.Cfg    `json:"grpc"`
	Address              string        `json:"address"`
	Advertise            string        `json:"advertise"`
	RegisterTTL          time.Duration `json:"register_ttl"`
	RegisterInterval     time.Duration `json:"register_interval"`
	SleepAfterDeRegister time.Duration `json:"sleepAfterDeRegister"`

	id       string
	name     string
	hostname string
}

func init() {
	transport.NewTCPServerTransport()
}
