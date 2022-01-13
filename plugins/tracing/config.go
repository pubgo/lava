package tracing

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/xerror"
)

var cfg = GetDefaultCfg()

type Cfg struct {
	Driver string `json:"driver"`
}

func (cfg Cfg) Build() (err error) {
	defer xerror.RespErr(&err)

	driver := cfg.Driver
	xerror.Assert(driver == "", "tracer driver is null")

	fc := GetFactory(driver)
	xerror.Assert(fc == nil, "tracer driver [%s] not found", driver)

	return fc(config.GetMap(Name))
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Driver: "noop",
	}
}
