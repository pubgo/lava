package entry

import (
	"context"

	"github.com/pubgo/lug/types"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Wrapper func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error
type Middleware func(next Wrapper) Wrapper

type Runtime interface {
	InitRT() error
	Start() error
	Stop() error
	Options() Opts
}

type Entry interface {
	AfterStop(func())
	BeforeStop(func())
	AfterStart(func())
	BeforeStart(func())
	OnCfg(fn interface{})
	Middleware(middleware Middleware)
	Description(description ...string)
	Flags(fn func(flags *pflag.FlagSet))
	Commands(commands ...*cobra.Command)
}

type Opt func(o *Opts)
type Opts struct {
	Name         string
	BeforeStarts []func()
	AfterStarts  []func()
	BeforeStops  []func()
	AfterStops   []func()
	Command      *cobra.Command
	Middlewares  []Middleware
}

func Watch(cb func(ent Entry)) error {
	return xerror.Wrap(dix.Provider(cb))
}
