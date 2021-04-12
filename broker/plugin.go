package broker

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/xerror"
)

func onInit(ent interface{}) {
	if !config.Decode(Name, &cfgList) {
		return
	}

	for name, cfg := range cfgList {
		brokers.Set(name, xerror.PanicErr(cfg.Build(name)).(Broker))
	}
}

func init() {
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
	})
}
