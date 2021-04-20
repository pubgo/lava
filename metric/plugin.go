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

	var reporter = xerror.PanicErr(cfg.Build()).(Reporter)
	SetDefault(reporter)
}

func init() {
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
	})

	vars.Watch(Name, func() interface{} {
		var dt map[string]Factory
		xerror.Panic(reporters.MapTo(&dt))
		return dt
	})
}
