package golug_entry

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type RunEntry interface {
	InitEntry() error
	Start() error
	Stop() error
	Options() Options
}

type Entry interface {
	Init(func())
	Run() RunEntry
	Version(v string)
	UnWrap(fn interface{})
	Dix(data ...interface{})
	Description(description ...string)
	Flags(fn func(flags *pflag.FlagSet))
	Commands(commands ...*cobra.Command)
}

type Option func(o *Options)
type Options struct {
	Init        func()
	Initialized bool
	Addr        string
	Name        string
	Version     string
	Command     *cobra.Command
}
