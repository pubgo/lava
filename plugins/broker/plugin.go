package broker

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func() {
			if config.Decode(Name, &cfgList) != nil {
				return
			}

			for name, cfg := range cfgList {
				var bk = xerror.PanicErr(cfg.Build(name)).(Broker)
				brokers.Set(name, bk)
				xerror.Exit(dix.ProviderNs(name, bk))
			}
		},
		OnVars: func(v types.Vars) {
			v(Name+"_factory", func() interface{} {
				var data = make(map[string]string)
				xerror.Panic(factories.Each(func(name string, fc Factory) {
					data[name] = stack.Func(fc)
				}))
				return data
			})

			v(Name, func() interface{} {
				var data = make(map[string]string)
				xerror.Panic(brokers.Each(func(name string, fc Broker) {
					data[name] = fc.String()
				}))
				return data
			})
		},
	})
}
