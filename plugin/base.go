package plugin

import (
	"fmt"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/lug/vars"

	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

var _ = Plugin(&Base{})

type Base struct {
	Name         string
	OnMiddleware func() types.Middleware
	OnInit       func(ent entry.Entry)
	OnCommands   func(cmd *cobra.Command)
	OnFlags      func(flags *pflag.FlagSet)
	OnWatch      func(name string, resp *types.WatchResp)
	OnVars       func(w func(name string, data func() interface{}))
	OnCodec      func(name string, resp *types.WatchResp) (map[string]interface{}, error)
}

func (p *Base) String() string { return p.Name }
func (p *Base) Init(ent entry.Entry) (err error) {
	return xerror.Try(func() {
		if p.OnMiddleware != nil {
			ent.Middleware(p.OnMiddleware())
		}

		if p.OnInit != nil {
			p.OnInit(ent)
		}

		if p.OnInit != nil && p.OnVars != nil {
			p.OnVars(vars.Watch)
		}
	})
}

func (p *Base) Watch(name string, r *types.WatchResp) (err error) {
	if p.OnWatch == nil {
		return
	}

	zap.L().Info(fmt.Sprintf("plugin [%s] watch", p.Name))
	return xerror.Try(func() { p.OnWatch(name, r) })
}

func (p *Base) Commands() *cobra.Command {
	if p.OnCommands == nil {
		return nil
	}

	cmd := &cobra.Command{Use: p.Name}
	xerror.TryThrow(func() { p.OnCommands(cmd) }, "commands callback error")
	return cmd
}

func (p *Base) Flags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(p.Name, pflag.PanicOnError)
	if p.OnFlags == nil {
		return flags
	}

	xerror.TryCatch(func() { p.OnFlags(flags) }, func(err error) {
		zap.L().Fatal("flags callback", zap.Any("err", err))
	})
	return flags
}
