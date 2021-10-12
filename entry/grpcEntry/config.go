package grpcEntry

import (
	"time"

	"github.com/pubgo/lava/logger"
	grpcGw "github.com/pubgo/lava/pkg/builder/grpc-gw"
	"github.com/pubgo/lava/pkg/builder/grpcs"

	"go.uber.org/zap"
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

var logs *zap.Logger

func init() {
	logs = logger.On(func(log *zap.Logger) { logs = log.Named(Name) })
}

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
