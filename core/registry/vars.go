package registry

import (
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/xerror"
)

func init() {
	vars.Register("list-service", func() interface{} {
		var srvList, err = Default().GetService(runmode.Project)
		xerror.Panic(err)
		return srvList
	})
}
