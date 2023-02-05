package registry

import (
	"github.com/pubgo/funk/vars"
	"github.com/pubgo/lava/core/runmode"
)

func init() {
	vars.Register("list-service", func() interface{} {
		return Default().GetService(runmode.Project)
	})
}
