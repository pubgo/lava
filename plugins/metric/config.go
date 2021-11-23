package metric

import "time"

var Name = "metric"

type Cfg struct {
	Driver    string        `json:"driver"`
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
