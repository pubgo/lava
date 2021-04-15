package tracer

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/xerror"
)

const Name = "tracer"
// error keys
const KeyErrorMessage        = "error_msg"
const KeyContextErrorMessage = "context_error_msg"

type Cfg struct {
	Driver string `json:"driver"`
}

func (cfg Cfg) Build() (_ opentracing.Tracer, err error) {
	defer xerror.RespErr(&err)

	driver := cfg.Driver
	xerror.Assert(driver == "", "tracer driver is null")

	fc := Get(driver)
	xerror.Assert(fc == nil, "tracer driver %s not found", driver)

	return fc(config.Map(Name))
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Driver: "jaeger",
	}
}
