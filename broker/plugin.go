package broker

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/xerror"
)

func onInit(ent interface{}) {
	config.Decode(Name, &cfgList)

	for name, cfg := range cfgList {
		driver := cfg.Driver
		xerror.Assert(driver == "", "broker driver is null")
		xerror.Assert(!factories.Has(driver), "broker driver %s not found", driver)

		fc := factories.Get(driver).(Factory)
		var brk = xerror.PanicErr(fc(config.Map(Name, name))).(Broker)

		xerror.Assert(brokers.Has(name), "broker %s driver %s already exists", name, driver)
		brokers.Set(name, brk)
	}
}

func init() {
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
	})
}
