package metric

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/vars"
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

	reporter := xerror.PanicErr(fc(config.Map(Name))).(Reporter)
	xerror.Assert(reporter == nil, "metric driver %s init error", driver)
	defaultReporter = reporter
}

func init() {
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
	})

	vars.Watch(Name, func() interface{} {
		var dt map[string]Factory
		xerror.Panic(reporters.Map(&dt))
		return dt
	})
}
