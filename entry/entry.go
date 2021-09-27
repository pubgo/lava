package entry

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/types"
)

type Runtime interface {
	InitRT() error
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

type Opt func(o *Opts)
type Opts struct {
	Name         string
	BeforeStarts []func()
	AfterStarts  []func()
	BeforeStops  []func()
	AfterStops   []func()
	Command      *cobra.Command
	Middlewares  []types.Middleware
}
