package orm

import "github.com/pubgo/lava/plugin"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			
		},
	})
}
