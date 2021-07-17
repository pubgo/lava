package entry

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Wrapper func(ctx context.Context, req Request, resp func(rsp interface{})) error
type Middleware func(next Wrapper) Wrapper

type Runtime interface {
	InitRT() error
	Start(args ...string) error
	Stop() error
	Options() Opts
}

type Entry interface {
	AfterStop(func())
	BeforeStop(func())
	AfterStart(func())
	BeforeStart(func())
	OnCfg(fn interface{})
	Dix(data ...interface{})
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
