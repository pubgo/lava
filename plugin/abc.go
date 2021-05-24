package plugin

import (
	"github.com/pubgo/lug/watcher"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const Name = "plugin"

type Manager interface {
	Plugins(...ManagerOpt) []Plugin
	Register(Plugin, ...ManagerOpt)
}

type ManagerOpt func(o *managerOpts)
type managerOpts struct {
	Module string
}

type Plugin interface {
	Init(ent interface{}) error
	Watch(name string, r *watcher.Response) error
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
