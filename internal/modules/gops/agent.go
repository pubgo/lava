package gops

import (
	"github.com/google/gops/agent"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/xerror"
)

func init() {
	if !runmode.IsDebug {
		return
	}

	xerror.Exit(agent.Listen(agent.Options{}))
}
