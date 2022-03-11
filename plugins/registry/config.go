package registry

import (
	"github.com/pubgo/xerror"
)

const DefaultPrefix = "/registry"

var Name = "registry"

type Cfg struct {
	Driver    string                 `json:"driver" yaml:"driver"`
	DriverCfg map[string]interface{} `json:"driver_config" yaml:"driver_config"`
}

func (cfg Cfg) Build() (_ Registry, err error) {
	defer xerror.RespErr(&err)

	var driver = cfg.Driver
	xerror.Assert(driver == "", "registry driver is null")
	xerror.Assert(!factories.Has(driver), "registry driver %s not found", driver)

	var fc = factories.Get(driver).(Factory)
	return fc(cfg.DriverCfg)
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Driver: "mdns",
	}
}
