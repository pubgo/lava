package entry

import (
	"github.com/pubgo/lug/plugin"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Runtime interface {
	Init() error
	Start() error
	Stop() error
	Options() Opts
}

type Entry interface {
	Version(v string)
	Dix(data ...interface{})
	OnCfg(fn interface{})
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
	Port         int
	Name         string
	Version      string
	Command      *cobra.Command
}
