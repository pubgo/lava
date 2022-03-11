package tracing

import (
	"github.com/pubgo/xerror"
)

var cfg = DefaultCfg()

type Cfg struct {
	Driver    string                 `json:"driver"`
	DriverCfg map[string]interface{} `json:"driver_config"`
}

func (cfg Cfg) Build() (err error) {
	defer xerror.RespErr(&err)

	driver := cfg.Driver
	xerror.Assert(driver == "", "tracer driver is null")

	fc := GetFactory(driver)
	xerror.Assert(fc == nil, "tracer driver [%s] not found", driver)

	return fc(cfg.DriverCfg)
}

func DefaultCfg() Cfg {
	return Cfg{
		Driver: "noop",
	}
}
