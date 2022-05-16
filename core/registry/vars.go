package registry

import (
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/xerror"
)

func init() {
	vars.Register("list-service", func() interface{} {
		var srv, err = Default().GetService(runtime.Project)
		xerror.Panic(err)
		return srv
	})
}
