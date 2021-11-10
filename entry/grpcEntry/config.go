package grpcEntry

import (
	"time"

	grpcGw "github.com/pubgo/lava/builder/grpc-gw"
	"github.com/pubgo/lava/builder/grpcs"
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
	Grpc                 *grpcs.Cfg    `json:"grpc"`
	Gw                   *grpcGw.Cfg   `json:"gw"`
	Address              string        `json:"address"`
	Advertise            string        `json:"advertise"`
	RegisterTTL          time.Duration `json:"register_ttl"`
	RegisterInterval     time.Duration `json:"register_interval"`
	SleepAfterDeRegister time.Duration `json:"sleepAfterDeRegister"`

	id       string
	name     string
	hostname string
}