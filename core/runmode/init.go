package runmode

import (
	"strconv"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"

	"github.com/pubgo/lava/internal/pkg/env"
)

var Mode = Local

func init() {
	defer recovery.Exit()

	mode := env.Get("lava_mode", "app_mode")
	if mode != "" {
		var i = assert.Must1(strconv.Atoi(mode))

		Mode = RunMode(i)
		assert.Assert(Mode.String() == "", "unknown mode, mode=%s", mode)
	}
}
