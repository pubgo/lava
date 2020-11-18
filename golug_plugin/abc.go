package golug_plugin

import (
	"github.com/pubgo/golug/golug_abc"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Manager is the internal_plugin manager which stores plugins and allows them to be retrieved.
// This is used by all the components of micro.
type Manager interface {
	Plugins(...ManagerOption) []Plugin
	Register(Plugin, ...ManagerOption) error
}

type ManagerOption func(o *ManagerOptions)
type ManagerOptions struct {
	Module string
}

type Response struct {
	Event    string
	Key      []byte
	Value    []byte
	Revision int64
}

// Plugin is the interface for plugins to micro. It differs from go-micro in that it's for
// the micro API, Web, Sidecar, CLI. It's a method of building middleware for the HTTP side.
type Plugin interface {
	Watch(r *Response) error
	Init(ent golug_abc.Entry) error
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
