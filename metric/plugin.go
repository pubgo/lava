package metric

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/consts"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/golug/tracelog"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func onInit(ent interface{}) {
	cfg := config.GetCfg().GetStringMap(Name)
	xerror.Assert(cfg == nil, "metric cfg is null")

	driver := cfg[consts.Driver]
	if driver == nil {
		driver = "prometheus"
		xlog.Warn("metric driver is null, set default prometheus")
	}

	fc := Get(driver.(string))
	xerror.Assert(fc == nil, "metric driver not found")

	delete(cfg, consts.Driver)

	defaultReporter = xerror.PanicErr(fc(cfg)).(Reporter)
	xerror.Assert(defaultReporter == nil, "metric driver %s init error", driver)
}

func init() {
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
	})

	tracelog.Watch(Name, func() interface{} { return List() })
}
