package cluster

import (
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

const Name = "cluster"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
		},
		OnVars: func(v types.Vars) {
		},
	})
}
