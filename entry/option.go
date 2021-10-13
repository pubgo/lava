package entry

import (
	"github.com/spf13/cobra"

	"github.com/pubgo/lava/types"
)

type Opt func(o *Opts)
type Opts struct {
	Name         string
	BeforeStarts []func()
	AfterStarts  []func()
	BeforeStops  []func()
	AfterStops   []func()
	Command      *cobra.Command
	Middlewares  []types.Middleware
}
