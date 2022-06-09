package runmode

import (
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/xerror"
	"strconv"
)

var Mode = Local

func init() {
	mode := env.Get("lava_mode", "app_mode")
	if mode != "" {
		var i, err = strconv.Atoi(mode)
		xerror.Panic(err)

		Mode = RunMode(i)
		xerror.Assert(Mode.String() == "", "unknown mode, mode=%s", mode)
	}
}
