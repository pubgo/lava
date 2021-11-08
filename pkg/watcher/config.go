package watcher

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
)

const Name = "watcher"

var cfg = DefaultCfg()

type Cfg struct {
	SkipNull bool   `json:"skip_null"`
	Driver   string `json:"driver"`
}

func (cfg Cfg) Build() (_ Watcher, err error) {
	defer xerror.RespErr(&err)

	driver := cfg.Driver
	xerror.Assert(driver == "", "watcher driver is null")
	xerror.Assert(factories[driver] == nil, "watcher driver [%s] not found", driver)

	fc := factories[driver]
	return fc(config.GetMap(Name))
}

func DefaultCfg() Cfg {
	return Cfg{
		Driver:   "noop",
		SkipNull: true,
	}
}
