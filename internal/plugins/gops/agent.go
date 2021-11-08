package gops

import (
	"github.com/google/gops/agent"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/xerror"
)

func init() {
	if runenv.IsProd() || runenv.IsRelease() {
		return
	}

	xerror.Exit(agent.Listen(agent.Options{}))
}
