package shutil

import "github.com/pubgo/lug/pkg/env"

var debug = false

func init() {
	env.GetBoolVal(&debug, "debug", "verbose")
}
