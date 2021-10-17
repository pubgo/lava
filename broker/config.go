package broker

import (
	"github.com/pubgo/lava/config"

	"github.com/pubgo/xerror"
)

var Name = "broker"
var cfgList = make(map[string]Cfg)

type Cfg struct {
	Driver string `json:"driver"`
}

func (cfg Cfg) Build(name string) (_ Broker, err error) {
	defer xerror.RespErr(&err)

	driver := cfg.Driver
	xerror.Assert(driver == "", "broker driver is null")
	xerror.Assert(!factories.Has(driver), "factory %s not found", driver)
	xerror.Assert(brokers.Has(name), "broker %s already exists", name)

	fc := factories.Get(driver).(Factory)
	return fc(config.GetMap(Name, name))
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Driver: "nsq",
	}
}
