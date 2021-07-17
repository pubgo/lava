package plugin

import (
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/lug/vars"

	"github.com/pubgo/x/try"
	"github.com/pubgo/xlog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

var _ = Plugin(&Base{})

type Base struct {
	Name         string
	OnMiddleware func() entry.Middleware
	OnInit       func(ent entry.Entry)
	OnWatch      func(name string, resp *types.Response)
	OnCodec      func(name string, resp *types.Response) (map[string]interface{}, error)
	OnCommands   func(cmd *cobra.Command)
	OnFlags      func(flags *pflag.FlagSet)
	OnVars       func(w func(name string, data func() interface{}))
}

func (p *Base) String() string { return p.Name }
func (p *Base) Init(ent entry.Entry) (err error) {
	return try.Try(func() {
		if p.OnMiddleware != nil {
			ent.Middleware(p.OnMiddleware())
		}

		if p.OnInit != nil {
			xlog.Infof("plugin [%s] init", p.Name)

			p.OnInit(ent)
		}

		if p.OnVars != nil {
			p.OnVars(vars.Watch)
		}
	})
}

func (p *Base) Watch(name string, r *types.Response) (err error) {
	return try.Try(func() {
		if p.OnWatch == nil {
			return
		}

		xlog.Infof("plugin [%s] watch", p.Name)
		p.OnWatch(name, r)
	})
}

func (p *Base) Commands() *cobra.Command {
	if p.OnCommands == nil {
		return nil
	}

	cmd := &cobra.Command{Use: p.Name}

	try.Catch(func() { p.OnCommands(cmd) }, func(err error) {
		xlog.Fatal("commands callback", zap.Any("err", err))
	})

	return cmd
}

func (p *Base) Flags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(p.Name, pflag.PanicOnError)
	if p.OnFlags != nil {
		try.Catch(func() { p.OnFlags(flags) }, func(err error) {
			xlog.Fatal("flags callback", zap.Any("err", err))
		})
	}
	return flags
}
