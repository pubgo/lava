package entry

import (
	"github.com/pubgo/lava/service/service_type"
	"github.com/urfave/cli/v2"
)

type Runtime interface {
	InitRT()
	Start() error
	Stop() error
	Options() Opts
	MiddlewareInter(middleware service_type.Middleware)
}

type Entry interface {
	AfterStop(func())
	BeforeStop(func())
	AfterStart(func())
	BeforeStart(func())
	Middleware(middleware service_type.Middleware)
	Description(description ...string)
	Flags(flags cli.Flag)
	Commands(commands *cli.Command)
}

type Handler interface {
	Init()
}
