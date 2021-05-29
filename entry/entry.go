package entry

import (
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/xerror"
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
	Version(v string)
	OnCfg(fn interface{})
	Dix(data ...interface{})
	Plugin(plugin plugin.Plugin)
	Description(description ...string)
	Flags(fn func(flags *pflag.FlagSet))
	Commands(commands ...*cobra.Command)
	BeforeStart(func())
	AfterStart(func())
	BeforeStop(func())
	AfterStop(func())
}

type Opt func(o *Opts)
type Opts struct {
	BeforeStarts []func()
	AfterStarts  []func()
	BeforeStops  []func()
	AfterStops   []func()
	Initialized  bool
	Name         string
	Version      string
	Command      *cobra.Command
}

func Parse(ent interface{}, fn func(ent Entry), errs ...func(b bool)) {
	xerror.Assert(ent == nil, "ent is nil")

	ent1, ok := ent.(Entry)
	if ok {
		fn(ent1)
	}

	if len(errs) > 0 {
		errs[0](ok)
	}
}
