package watcher

import (
	"github.com/pubgo/xerror"
)

const Name = "watcher"

var cfg = DefaultCfg()

type Cfg struct {
	SkipNull bool                                        `json:"skip_null"`
	Driver   string                                      `json:"driver"`
	Projects []string                                    `json:"projects"`
	Set      func(string, interface{})                   `json:"-" yaml:"-"`
	GetMap   func(keys ...string) map[string]interface{} `json:"-" yaml:"-"`
}

func (cfg Cfg) Build(data map[string]interface{}) (_ Watcher, err error) {
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
