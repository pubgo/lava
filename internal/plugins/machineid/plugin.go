package machineid

import (
	"github.com/denisbrodbeck/machineid"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/plugin"
)

// Name deviceId
const Name = "machineid"

var logs = logger.Component(Name)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			// TODO hook
			id, err := machineid.ID()
			xerror.Panic(err)
			logs.S().Infow("machineid", "value", id)
		},
	})
}
