package metric

import (
	"time"

	"github.com/pubgo/lava/core/config"
)

var Name = "metric"

type Config struct {
	Driver    string        `json:"driver"`
	DriverCfg config.Map    `json:"driver_config"`
	Interval  time.Duration `json:"interval"`
}

func DefaultCfg() Config {
	return Config{
		Driver:   "noop",
		Interval: 2 * time.Second,
	}
}
