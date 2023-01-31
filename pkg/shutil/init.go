package shutil

import (
	"github.com/pubgo/funk/env"
)

var debug = false

func init() {
	env.GetBoolVal(&debug, "debug", "verbose")
}
