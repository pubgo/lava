package machineid

import (
	"github.com/denisbrodbeck/machineid"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/plugin"
)

const Name = "machineid"

var logs = logz.Component(Name)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			//TODO hook
			id, err := machineid.ID()
			xerror.Panic(err)
			logs.Infow("machineid", "value", id)
		},
	})
}
