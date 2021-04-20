package metric

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/xerror"
)

var Name = "metric"

type Cfg struct {
	Driver string `json:"driver"`
}

func (cfg Cfg) Build() (_ Reporter, err error) {
	defer xerror.RespErr(&err)

	driver := cfg.Driver
	xerror.Assert(driver == "", "metric driver is null")

	fc := Get(driver)
	xerror.Assert(fc == nil, "metric driver %s not found", driver)

	return fc(config.GetMap(Name))
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Driver: "noop",
	}
}
