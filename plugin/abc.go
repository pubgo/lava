package plugin

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/pubgo/lug/types"
)

const Name = "plugin"

type Opt func(o *options)
type options struct {
	Module string
}

type Entry interface {
	AfterStop(func())
	BeforeStop(func())
	AfterStart(func())
	BeforeStart(func())
}

type Plugin interface {
	String() string
	Id() string
	Flags() *pflag.FlagSet
	Commands() *cobra.Command
	Init(ent Entry) error
	Watch(name string, r *types.WatchResp) error
	Vars(func(name string, data func() interface{})) error
	Health() func(ctx context.Context) error
	Middleware() types.Middleware
}
