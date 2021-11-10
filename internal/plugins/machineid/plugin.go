package machineid

import (
	"github.com/denisbrodbeck/machineid"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/internal/logz"
	"github.com/pubgo/lava/plugin"
)

const Name = "machineid"

var logs = logz.New(Name)

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