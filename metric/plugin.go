package metric

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/golug/vars"
	"github.com/pubgo/xerror"
)

func onInit(ent interface{}) {
	var cfg = GetDefaultCfg()
	if !config.Decode(Name, &cfg) {
		return
	}

	driver := cfg.Driver
	xerror.Assert(driver == "", "metric driver is null")

	fc := Get(driver)
	xerror.Assert(fc == nil, "metric driver %s not found", driver)

	defaultReporter = xerror.PanicErr(fc(config.Map(Name))).(Reporter)
	xerror.Assert(defaultReporter == nil, "metric driver %s init error", driver)
}

func init() {
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
	})

	vars.Watch(Name, func() interface{} { return List() })
}
