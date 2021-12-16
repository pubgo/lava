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
			id, err := machineid.ID()
			xerror.Panic(err)
			logs.Infof("machineid=>%s", id)
		},
	})
}
