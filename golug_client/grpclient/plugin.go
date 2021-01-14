package grpclient

import (
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
)

func init() {
	golug_plugin.Register(&golug_plugin.Base{
		Name: Name,
		OnInit: func(ent golug_entry.Entry) {
			golug_config.Decode(Name, &cfg)
			initOption(cfg)
		},
	})
}
