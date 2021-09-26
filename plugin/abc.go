package plugin

import (
	"context"

	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/types"

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
	Init(ent entry.Entry) error
	Watch(name string, r *types.WatchResp) error
	Vars(func(name string, data func() interface{})) error
	Health() func(ctx context.Context) error
	Middleware() types.Middleware
}
