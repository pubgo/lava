package gops

import (
	"github.com/google/gops/agent"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/runtime"
)

func init() {
	if runtime.IsProd() || runtime.IsRelease() {
		return
	}

	xerror.Exit(agent.Listen(agent.Options{}))
}
