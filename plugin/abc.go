package plugin

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/pubgo/lava/types"
)

const Name = "plugin"

type Plugin interface {
	String() string
	UniqueName() string
	Flags() *pflag.FlagSet
	Commands() *cobra.Command
	Init() error
	Watch() func(name string, r *types.WatchResp) error
	Vars(func(name string, data func() interface{})) error
	Health() func(ctx context.Context) error
	Middleware() types.Middleware
}
