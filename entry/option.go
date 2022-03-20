package entry

import (
	"github.com/pubgo/lava/service/service_type"
	"github.com/urfave/cli/v2"
)

type Opt func(o *Opts)
type Opts struct {
	Name         string
	BeforeStarts []func()
	AfterStarts  []func()
	BeforeStops  []func()
	AfterStops   []func()
	Command      *cli.Command
	Handlers     []Handler
	Middlewares  []service_type.Middleware
}
