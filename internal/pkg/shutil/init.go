package shutil

import (
	"github.com/pubgo/lava/internal/pkg/env"
)

var debug = false

func init() {
	env.GetBoolVal(&debug, "debug", "verbose")
}
