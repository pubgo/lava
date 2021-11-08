package plugin

import (
	"context"

	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/types"
)

const Name = "plugin"

type Plugin interface {
	String() string
	UniqueName() string
	Flags() []cli.Flag
	Commands() *cli.Command
	Init() error
	Watch() func(name string, r *types.WatchResp) error
	Vars(func(name string, data func() interface{})) error
	Health() func(ctx context.Context) error
	Middleware() types.Middleware
}
