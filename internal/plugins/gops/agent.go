package gops

import (
	"github.com/google/gops/agent"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/xerror"
)

func init() {
	if runtime.IsProd() || runtime.IsRelease() {
		return
	}

	xerror.Exit(agent.Listen(agent.Options{}))
}
