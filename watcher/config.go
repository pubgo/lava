package watcher

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/xerror"
)

const Name = "watcher"

var cfg = DefaultCfg()

type Cfg struct {
	SkipNull bool   `json:"skip_null"`
	Driver   string `json:"driver"`

	// Projects 需要watcher的项目
	Projects []string `json:"projects"`

	cfg config.Config
}

func (cfg Cfg) Build(data types.M) (_ Watcher, err error) {
	defer xerror.RespErr(&err)

	driver := cfg.Driver
	xerror.Assert(driver == "", "watcher driver is null")
	xerror.Assert(factories[driver] == nil, "watcher driver [%s] not found", driver)

	fc := factories[driver]
	return fc(data)
}

func DefaultCfg() Cfg {
	return Cfg{
		Driver:   "noop",
		SkipNull: true,
	}
}
