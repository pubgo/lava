package grpclient

import (
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
)

func init() {
	xerror.Panic(golug_plugin.Register(&golug_plugin.Base{
		Name: Name,
		OnInit: func(ent golug_entry.Entry) {
			xerror.Panic(ent.Decode(Name, &cfg))

			for k, v := range cfg.Configs {
				if !v.Enabled {
					continue
				}

				initClient(k, v)
			}
		},
	}))
}