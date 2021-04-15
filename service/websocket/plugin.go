package websocket

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: name,
		OnInit: func(ent interface{}) {

			config.Decode(name, &cfg)
		},
	})
}
