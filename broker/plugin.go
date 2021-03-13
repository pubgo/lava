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
		xerror.Assert(brokers.Has(name), "broker %s already exists", name)

		fc := factories.Get(driver).(Factory)
		brokers.Set(name, xerror.PanicErr(fc(config.Map(Name, name))).(Broker))
	}
}

func init() {
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
	})
}
