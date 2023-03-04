package metric

import (
	"time"

	"github.com/pubgo/lava/core/config"
)

var Name = "metric"

type Cfg struct {
	Driver    string        `json:"driver"`
	DriverCfg config.CfgMap `json:"driver_config"`
	Interval  time.Duration `json:"interval"`
}

func DefaultCfg() Cfg {
	return Cfg{
		Driver:   "noop",
		Interval: 2 * time.Second,
	}
}
