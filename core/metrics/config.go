package metrics

import (
	"github.com/pubgo/funk/config"
	"time"
)

var Name = "metric"

type MetricConfigLoader struct {
	Metric *Config `yaml:"metric"`
}

type Config struct {
	Driver    string        `yaml:"driver"`
	DriverCfg *config.Node  `yaml:"driver_config"`
	Interval  time.Duration `yaml:"interval"`
}

func DefaultCfg() Config {
	return Config{
		Driver:   "noop",
		Interval: 2 * time.Second,
	}
}
