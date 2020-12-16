package golug_entry

import (
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
	Run() RunEntry
	Version(v string)
	UnWrap(fn interface{}) error
	Dix(data ...interface{}) error
	Description(description ...string)
	Flags(fn func(flags *pflag.FlagSet))
	Commands(commands ...*cobra.Command)
}

type Option func(o *Options)
type Options struct {
	Initialized bool
	Addr        string
	Name        string
	Version     string
	Command     *cobra.Command
}
