package gops

import (
	"github.com/google/gops/agent"
	"github.com/pubgo/xerror"
)

func init() {
	xerror.Exit(agent.Listen(agent.Options{}))
}
