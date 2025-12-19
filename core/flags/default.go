package flags

import (
	"github.com/pubgo/funk/v2/running"
)

func init() {
	Register(running.DebugFlag)
	Register(running.EnvFlag)
	Register(running.ConfFlag)
	Register(running.GrpcPortFlag)
	Register(running.HttpPortFlag)
}
