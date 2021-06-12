package plugin

import (
	"github.com/pubgo/lug/watcher"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const Name = "plugin"

type Opt func(o *options)
type options struct {
	Module string
}

type Plugin interface {
	String() string
	Flags() *pflag.FlagSet
	Commands() *cobra.Command
	Init(ent interface{}) error
	Watch(name string, r *watcher.Response) error
}
