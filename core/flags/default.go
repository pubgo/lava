package flags

import (
	"github.com/pubgo/funk/v2/config/configflags"

	"github.com/pubgo/lava/v2/core/running"
)

func init() {
	Register(running.DebugFlag)
	Register(running.EnvFlag)
	Register(configflags.ConfFlag)
	Register(running.GrpcPortFlag)
	Register(running.HttpPortFlag)
}
