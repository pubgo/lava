package gops

import (
	"github.com/google/gops/agent"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/lava/core/runmode"
)

func init() {
	if !runmode.IsDebug {
		return
	}

	assert.Exit(agent.Listen(agent.Options{}))
}
