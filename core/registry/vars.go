package registry

import (
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/xerror"
)

func init() {
	vars.Register("list-service", func() interface{} {
		var srv, err = Default().GetService(runmode.Project)
		xerror.Panic(err)
		return srv
	})
}
