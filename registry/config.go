package registry

import (
	"github.com/pubgo/golug/types"
)

var Name = "registry"
var cfg = GetDefaultCfg()

type Cfg struct {
	Project string         `json:"project"`
	Driver  string         `json:"driver"`
	Name    string         `json:"name"`
	Prefix  string         `json:"prefix"`
	TTL     types.Duration `json:"ttl"`
	Timeout types.Duration `json:"timeout"`
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Prefix: "/registry",
		Driver: "mdns",
	}
}
