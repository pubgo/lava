package registry

import (
	"github.com/pubgo/funk/runmode"
	"github.com/pubgo/lava/core/vars"
)

func init() {
	vars.Register("list-service", func() interface{} {
		return Default().GetService(runmode.Project)
	})
}
