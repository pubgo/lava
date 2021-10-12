package plugin

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/pubgo/lava/internal/logs"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var _ json.Marshaler = (*Base)(nil)
var _ Plugin = (*Base)(nil)

type Base struct {
	Name         string
	Descriptor   string
	Url          string
	OnHealth     func(ctx context.Context) error
	OnMiddleware types.Middleware
	OnInit       func(ent Entry)
	OnCommands   func(cmd *cobra.Command)
	OnFlags      func(flags *pflag.FlagSet)
	OnWatch      func(name string, resp *types.WatchResp)
	OnVars       func(w func(name string, data func() interface{}))
}

func (p *Base) getFuncStack(val interface{}) string {
	r := reflect.ValueOf(val)
	if !r.IsValid() || r.IsNil() {
		return ""
	}
	return stack.Func(val)
}

func (p *Base) MarshalJSON() ([]byte, error) {
	defer xerror.RespRaise()
	var data = make(map[string]string)
	data["name"] = p.Name
	data["descriptor"] = p.Descriptor
	data["url"] = p.Url
	data["health"] = p.getFuncStack(p.OnHealth)
	data["middleware"] = p.getFuncStack(p.OnMiddleware)
	data["init"] = p.getFuncStack(p.OnInit)
	data["commands"] = p.getFuncStack(p.OnCommands)
	data["flags"] = p.getFuncStack(p.OnFlags)
	data["watch"] = p.getFuncStack(p.OnWatch)
	data["vars"] = p.getFuncStack(p.OnVars)
	return json.Marshal(data)
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

func (p *Base) Middleware() types.Middleware { return p.OnMiddleware }
func (p *Base) String() string               { return p.Descriptor }
func (p *Base) Id() string                   { return p.Name }
func (p *Base) Init(ent Entry) error {
	if p.OnInit == nil {
		return nil
	}

	return xerror.Try(func() { p.OnInit(ent) })
}

func (p *Base) Watch(name string, r *types.WatchResp) (err error) {
	if p.OnWatch == nil {
		return
	}

	logs.Named(Name).Infof("plugin [%s] watch init", p.Name)
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

	xerror.TryThrow(func() { p.OnFlags(flags) }, "flags callback")
	return flags
}
