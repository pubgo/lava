package metric

import (
	"github.com/pubgo/lava/config"
	"time"
)

var Name = "metric"

type Cfg struct {
	Driver    string        `json:"driver"`
	DriverCfg config.CfgMap `json:"driver_config"`
	Interval  time.Duration `json:"interval"`
	Separator string        `json:"separator"`
}

func DefaultCfg() Cfg {
	return Cfg{
		Driver:    "noop",
		Interval:  time.Second,
		Separator: "_",
	}
}
