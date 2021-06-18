package grpc

import (
	grpcWeb "github.com/pubgo/lug/builder/grpc-web"
	"github.com/pubgo/lug/builder/grpcs"

	"time"
)

type Cfg struct {
	Srv                  *grpcs.Cfg     `json:"grpc"`
	Web                  *grpcWeb.Cfg   `json:"web"`
	SleepAfterDeregister time.Duration `json:"sleepAfterDeregister"`
	RegisterInterval     time.Duration `json:"registerInterval"`
	RegisterTTL          time.Duration `json:"register_ttl"`
	Address              string        `json:"address"`
	Advertise            string        `json:"advertise"`
	hostname             string
	id                   string
	name                 string
}

const (
	// DefaultMaxMsgSize define maximum message size that server can send or receive.
	// Default value is 4MB.
	DefaultMaxMsgSize = 1024 * 1024 * 4

	DefaultSleepAfterDeregister = time.Second * 2

	// DefaultRegisterTTL The register expiry time
	DefaultRegisterTTL = time.Minute

	// DefaultRegisterInterval The interval on which to register
	DefaultRegisterInterval = time.Second * 30
)
