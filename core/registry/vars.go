package registry

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/vars"
)

func init() {
	defer recovery.Exit()

	vars.Register("list-service", func() interface{} {
		return assert.Must1(Default().GetService(runmode.Project))
	})
}
