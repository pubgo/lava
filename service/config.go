package service

import (
	"github.com/pubgo/lava/service/internal/fiber_builder"
	"time"

	"github.com/pubgo/lava/entry/grpcEntry/grpcs"
)

const Name = "grpc_entry"

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
	GrpcWeb   bool      `json:"grpc_web"`
	Grpc      grpcs.Cfg `json:"grpc"`
	Gw        fiber_builder.Cfg
	Address   string `json:"address"`
	Advertise string `json:"advertise"`

	id       string
	name     string
	hostname string
}
