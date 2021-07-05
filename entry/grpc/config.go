package grpc

import (
	grpcGw "github.com/pubgo/lug/builder/grpc-gw"
	"github.com/pubgo/lug/builder/grpcs"

	"github.com/pubgo/xlog"
	_ "github.com/pubgo/xlog/xlog_grpc"

	"time"
)

const Name = "grpc_entry"

var logs = xlog.GetLogger(Name)

type Cfg struct {
	Rpc                  *grpcs.Cfg    `json:"grpc"`
	Gw                   *grpcGw.Cfg   `json:"gw"`
	SleepAfterDeRegister time.Duration `json:"sleepAfterDeRegister"`
	RegisterInterval     time.Duration `json:"register_interval"`
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

	DefaultSleepAfterDeRegister = time.Second * 2

	// DefaultRegisterTTL The register expiry time
	DefaultRegisterTTL = time.Minute

	// DefaultRegisterInterval The interval on which to register
	DefaultRegisterInterval = time.Second * 30
)
