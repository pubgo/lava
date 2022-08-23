package registry

import (
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/vars"
)

func init() {
	defer recovery.Exit()

	vars.Register("list-service", func() interface{} {
		var r = Default().GetService(runmode.Project).ToResult()
		if r.IsErr() {
			return r.Err()
		}
		return r.Unwrap()
	})
}
