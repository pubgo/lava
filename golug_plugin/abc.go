package golug_plugin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/golug/golug_types"
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

// Plugin is the interface for plugins to micro. It differs from go-micro in that it's for
// the micro API, Web, Sidecar, CLI. It's a method of building middleware for the HTTP side.
type Plugin interface {
	Watch(r golug_types.CfgValue) error
	Init(r golug_types.CfgValue) error
	Flags() *pflag.FlagSet
	Commands() *cobra.Command
	Handler() fiber.Handler
	String() string
}

type Option func(o *Options)
type Options struct {
	Name     string
	Flags    *pflag.FlagSet
	Commands *cobra.Command
}
