package tracing

import (
	"github.com/pubgo/lug/config"

	"github.com/pubgo/xerror"
)

const Name = "tracing"

type Cfg struct {
	Driver string `json:"driver"`
}

func (cfg Cfg) Build() (_ Tracer, err error) {
	defer xerror.RespErr(&err)

	driver := cfg.Driver
	xerror.Assert(driver == "", "tracer driver is null")

	fc := Get(driver)
	xerror.Assert(fc == nil, "tracer driver [%s] not found", driver)

	return fc(config.GetMap(Name))
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Driver: "jaeger",
	}
}
