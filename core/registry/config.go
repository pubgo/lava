package registry

import (
	"fmt"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
)

// https://github.com/go-eagle/eagle/blob/master/pkg/registry/registry.go

const DefaultPrefix = "/registry"

var Name = "registry"

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

type Config struct {
	RegisterInterval time.Duration          `yaml:"registerInterval"`
	Driver           string                 `json:"driver" yaml:"driver"`
	DriverCfg        map[string]interface{} `json:"driver_config" yaml:"driver_config"`
}

func (cfg *Config) Check() *Config {
	assert.Fn(cfg.Driver == "", func() error {
		err := fmt.Errorf("registry driver is null")
		return errors.WrapKV(err, "cfg", cfg)
	})
	return cfg
}

func DefaultCfg() Config {
	return Config{
		Driver: "mdns",
	}
}
