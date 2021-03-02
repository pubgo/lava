package grpclient

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent interface{}) {
			var cfg = GetDefaultCfg()
			config.Decode(Name, &cfg)
		},
	})
}
