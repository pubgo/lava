package entry

import (
	"github.com/pubgo/lug/types"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Runtime interface {
	InitRT() error
	Start() error
	Stop() error
	Options() Opts
}

type Entry interface {
	AfterStop(func())
	BeforeStop(func())
	AfterStart(func())
	BeforeStart(func())
	OnCfg(fn interface{})
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
