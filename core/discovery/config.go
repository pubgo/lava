package discovery

import (
	"fmt"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
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
