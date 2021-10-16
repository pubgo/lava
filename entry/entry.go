package entry

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

type Runtime interface {
	InitRT()
	Start() error
	Stop() error
	Options() Opts
	MiddlewareInter(middleware types.Middleware)
}

type Entry interface {
	AfterStop(func())
	BeforeStop(func())
	AfterStart(func())
	BeforeStart(func())
	Plugin(plugins ...plugin.Plugin)
	Middleware(middleware types.Middleware)
	Description(description ...string)
	Flags(fn func(flags *pflag.FlagSet))
	Commands(commands ...*cobra.Command)
}

type InitHandler interface {
	Init()
}

type WatchHandler interface {
	Watch() (name string)
}
