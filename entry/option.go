package entry

import (
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/types"
)

type Opt func(o *Opts)
type Opts struct {
	Name          string
	BeforeStarts  []func()
	AfterStarts   []func()
	BeforeStops   []func()
	AfterStops    []func()
	Command       *cli.Command
	Handlers      []InitHandler
	WatchProjects []string
	Middlewares   []types.Middleware
}
