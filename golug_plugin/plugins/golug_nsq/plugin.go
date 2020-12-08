package golug_nsq

import (
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
)

func init() {
	xerror.Panic(golug_plugin.Register(&golug_plugin.Base{
		Name: name,
		OnInit: func(ent golug_entry.Entry) {

			xerror.Panic(ent.Decode(name, &cfg))

			for k, v := range cfg.Configs {
				if !v.Enabled {
					continue
				}

				initClient(k, v)
			}
		},
	}))
}
