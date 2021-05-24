package watcher

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/xerror"

	"strings"
)

const Name = "watcher"

type Cfg struct {
	Prefix   string   `json:"prefix"`
	Driver   string   `json:"driver"`
	Projects []string `json:"projects"`
}

func (cfg Cfg) Build() (_ Watcher, err error) {
	defer xerror.RespErr(&err)

	driver := cfg.Driver
	xerror.Assert(driver == "", "watcher driver is null")
	xerror.Assert(factories[driver] == nil, "watcher driver [%s] not found", driver)

	fc := factories[driver]
	return fc(config.GetCfg().GetStringMap(Name))
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Prefix: "/watcher",
		Driver: "etcd",
	}
}

//  /projectName/foo/bar -->  projectName.foo.bar
func KeyToDot(prefix string) string {
	return strings.Trim(strings.ReplaceAll(prefix, "/", "."), ".")
}
