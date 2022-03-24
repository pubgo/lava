package cluster

import (
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/vars"
)

const Name = "cluster"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
		},
		OnVars: func(v vars.Publisher) {
		},
	})
}
