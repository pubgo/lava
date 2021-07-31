package watcher

import (
	"strings"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/runenv"

	"github.com/pubgo/xerror"
)

const Name = "watcher"

var cfg = GetDefaultCfg()

type Cfg struct {
	SkipNull bool     `json:"skip_null"`
	Driver   string   `json:"driver"`
	Projects []string `json:"projects"`
	Exclude  []string `json:"exclude"`
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
		Driver:   "etcd",
		SkipNull: true,
	}
}

func trimProject(key string) string {
	return strings.Trim(strings.TrimPrefix(key, runenv.Project), ".")
}

// KeyToDot /projectName/foo/bar -->  projectName.foo.bar
func KeyToDot(prefix ...string) string {
	var p string
	if len(prefix) > 0 {
		p = strings.Join(prefix, ".")
	}

	p = strings.ReplaceAll(strings.ReplaceAll(p, "/", "."), "..", ".")
	p = strings.Trim(p, ".")

	return p
}
