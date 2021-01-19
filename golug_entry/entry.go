package golug_entry

import (
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type RunEntry interface {
	Init() error
	Start() error
	Stop() error
	Options() Options
}

type Entry interface {
	Version(v string)
	Dix(data ...interface{})
	OnCfg(fn interface{})
	Plugin(plugin golug_plugin.Plugin)
	Description(description ...string)
	Flags(fn func(flags *pflag.FlagSet))
	Commands(commands ...*cobra.Command)
	WithBeforeStart(func(_ *BeforeStart))
	WithAfterStart(func(_ *AfterStart))
	WithBeforeStop(func(_ *BeforeStop))
	WithAfterStop(func(_ *AfterStop))
}

type Option func(o *Options)
type Options struct {
	Initialized bool
	Port        uint
	Name        string
	Version     string
	Command     *cobra.Command
}

type BeforeStart = dix_run.BeforeStartCtx
type BeforeStop = dix_run.BeforeStopCtx
type AfterStart = dix_run.AfterStartCtx
type AfterStop = dix_run.AfterStopCtx
