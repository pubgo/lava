package registry

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/xerror"
)

func onInit(ent interface{}) {
	var cfg = GetDefaultCfg()
	if !config.Decode(Name, &cfg) {
		return
	}

	var driver = cfg.Driver
	xerror.Assert(driver == "", "registry driver is null")
	xerror.Assert(!factories.Has(driver), "registry driver %s not found", driver)

	var fc = factories.Get(driver).(Factory)
	Default = xerror.PanicErr(fc(config.Map(Name))).(Registry)
}

func init() {
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
	})
}
