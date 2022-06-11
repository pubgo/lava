package runmode

import (
	"strconv"

	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/internal/pkg/env"
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
