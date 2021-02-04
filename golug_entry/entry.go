package golug_entry

import (
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
	BeforeStart(func())
	AfterStart(func())
	BeforeStop(func())
	AfterStop(func())
}

type Option func(o *Options)
type Options struct {
	Initialized bool
	Port        uint
	Name        string
	Version     string
	Command     *cobra.Command
}
