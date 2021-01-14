package golug_registry

import (
	"time"
)

var Name = "registry"
var defaultPrefix = "/registry"
var cfg Cfg

type Cfg struct {
	Project string        `json:"project"`
	Driver  string        `json:"driver"`
	Name    string        `json:"name"`
	Prefix  string        `json:"prefix"`
	TTL     time.Duration `json:"ttl"`
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Prefix: defaultPrefix,
	}
}
