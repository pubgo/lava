package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"reflect"
)

var _ json.Marshaler = (*Base)(nil)
var _ Plugin = (*Base)(nil)

type Base struct {
	Name         string
	OnHealth     func(ctx context.Context) error
	OnMiddleware types.Middleware
	OnInit       func(ent entry.Entry)
	OnCommands   func(cmd *cobra.Command)
	OnFlags      func(flags *pflag.FlagSet)
	OnWatch      func(name string, resp *types.WatchResp)
	OnVars       func(w func(name string, data func() interface{}))
}

func (p *Base) getStack(val interface{}) string {
	if !reflect.ValueOf(val).IsValid() || reflect.ValueOf(val).IsNil() {
		return ""
	}
	return stack.Func(val)
}

func (p *Base) MarshalJSON() ([]byte, error) {
	var data = make(map[string]string)
	data["name"] = p.Name
	data["health"] = p.getStack(p.OnHealth)
	data["middleware"] = p.getStack((func(next types.MiddleNext) types.MiddleNext)(p.OnMiddleware))
	data["init"] = p.getStack(p.OnInit)
	data["commands"] = p.getStack(p.OnCommands)
	data["flags"] = p.getStack(p.OnFlags)
	data["watch"] = p.getStack(p.OnWatch)
	data["vars"] = p.getStack(p.OnVars)
	return json.Marshal(data)
}

func (p *Base) Middleware() types.Middleware {
	return p.OnMiddleware
}

func (p *Base) Vars(f func(name string, data func() interface{})) error {
	if p.OnVars == nil {
		return nil
	}

	return xerror.Try(func() { p.OnVars(f) })
}

func (p *Base) Health() func(ctx context.Context) error {
	if p.OnHealth == nil {
		return func(ctx context.Context) error { return nil }
	}
	return p.OnHealth
}

func (p *Base) String() string { return p.Name }
func (p *Base) Init(ent entry.Entry) error {
	if p.OnInit == nil {
		return nil
	}

	return xerror.Try(func() { p.OnInit(ent) })
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
