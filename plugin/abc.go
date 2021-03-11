package plugin

import (
	"github.com/pubgo/golug/watcher"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const Name ="plugin"
type Manager interface {
	Plugins(...ManagerOpt) []Plugin
	Register(Plugin, ...ManagerOpt)
}

type ManagerOpt func(o *ManagerOpts)
type ManagerOpts struct {
	Module string
}

type Plugin interface {
	Watch(r *watcher.Response) error
	Init(ent interface{}) error
	Flags() *pflag.FlagSet
	Commands() *cobra.Command
	String() string
}

type Option func(o *Options)
type Options struct {
	Name     string
	Flags    *pflag.FlagSet
	Commands *cobra.Command
}
