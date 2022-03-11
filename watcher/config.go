package watcher

import (
	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/xerror"
)

type Cfg struct {
	SkipNull bool   `json:"skip_null"`
	Driver   string `json:"driver"`

	// Projects 需要watcher的项目
	Projects []string `json:"projects"`

	cfg config_type.Interface
}

func (cfg Cfg) Build(data types.CfgMap) (_ Watcher, err error) {
	defer xerror.RespErr(&err)

	driver := cfg.Driver
	xerror.Assert(driver == "", "watcher driver is null")
	xerror.Assert(GetFactory(driver) == nil, "watcher driver [%s] not found", driver)

	fc := GetFactory(driver)
	return fc(data)
}

func DefaultCfg() Cfg {
	return Cfg{
		Driver:   "noop",
		SkipNull: true,
	}
}
