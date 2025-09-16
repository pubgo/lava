package discovery

import (
	"fmt"
	"time"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/errors"
)

// https://github.com/go-eagle/eagle/blob/master/pkg/registry/registry.go
// https://github.com/prometheus/prometheus/tree/main/discovery

type Config struct {
	Interval  time.Duration `yaml:"interval"`
	Driver    string        `yaml:"driver"`
	DriverCfg *config.Node  `yaml:"driver_config"`
}

func (cfg *Config) Check() *Config {
	assert.Fn(cfg.Driver == "", func() error {
		err := fmt.Errorf("registry driver is null")
		return errors.WrapKV(err, "cfg", cfg)
	})
	return cfg
}
