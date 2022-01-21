package plugin

import (
	"encoding/json"
	"reflect"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/types"
)

var _ json.Marshaler = (*Base)(nil)
var _ Plugin = (*Base)(nil)

type Base struct {
	Name         string
	Descriptor   string
	Url          string
	Docs         interface{}
	OnHealth     types.Healthy
	OnMiddleware types.Middleware
	OnInit       func(p Process)
	OnCommands   func() *types.Command
	OnFlags      func() types.Flags
	OnWatch      types.Watcher
	OnVars       func(v types.Vars)

	beforeStarts []func()
	afterStarts  []func()
	beforeStops  []func()
	afterStops   []func()
}

func (p *Base) BeforeStart(fn func())  { p.beforeStarts = append(p.beforeStarts, fn) }
func (p *Base) AfterStart(fn func())   { p.afterStarts = append(p.afterStarts, fn) }
func (p *Base) BeforeStop(fn func())   { p.beforeStops = append(p.beforeStops, fn) }
func (p *Base) AfterStop(fn func())    { p.afterStops = append(p.afterStops, fn) }
func (p *Base) BeforeStarts() []func() { return p.beforeStarts }
func (p *Base) AfterStarts() []func()  { return p.afterStarts }
func (p *Base) BeforeStops() []func()  { return p.beforeStops }
func (p *Base) AfterStops() []func()   { return p.afterStops }

// getFuncStack 获取函数stack信息
func (p *Base) getFuncStack(val interface{}) string {
	r := reflect.ValueOf(val)
	if !r.IsValid() || r.IsNil() {
		return ""
	}
	return stack.Func(val)
}

func (p *Base) MarshalJSON() ([]byte, error) {
	defer xerror.RespRaise()
	var data = make(map[string]interface{})
	data["name"] = p.Name
	data["docs"] = p.Docs
	data["descriptor"] = p.Descriptor
	data["url"] = p.Url
	data["health"] = p.getFuncStack(p.OnHealth)
	data["middleware"] = p.getFuncStack(p.OnMiddleware)
	data["init"] = p.getFuncStack(p.OnInit)
	data["commands"] = p.getFuncStack(p.OnCommands)
	data["flags"] = p.getFuncStack(p.OnFlags)
	data["watch"] = p.getFuncStack(p.OnWatch)
	data["expvar"] = p.getFuncStack(p.OnVars)

	var handler = func(fns []func()) (data []string) {
		for i := range fns {
			data = append(data, stack.Func(fns[i]))
		}
		return
	}

	data["beforeStarts"] = handler(p.beforeStarts)
	data["beforeStops"] = handler(p.beforeStops)
	data["afterStarts"] = handler(p.afterStarts)
	data["afterStops"] = handler(p.afterStops)
	return json.Marshal(data)
}

func (p *Base) Vars(f types.Vars) error {
	if p.OnVars == nil {
		return nil
	}

	return xerror.Try(func() { p.OnVars(f) })
}

func (p *Base) Health() types.Healthy {
	if p.OnHealth == nil {
		return nil
	}

	return p.OnHealth
}

func (p *Base) Middleware() types.Middleware { return p.OnMiddleware }
func (p *Base) String() string               { return p.Descriptor }
func (p *Base) ID() string                   { return p.Name }
func (p *Base) Init() (gErr error) {
	defer xerror.Resp(func(err xerror.XErr) {
		gErr = err.WrapF("plugin: %s", p.Name)
	})

	if p.OnInit == nil {
		return nil
	}

	p.OnInit(p)
	return nil
}

func (p *Base) Watch() types.Watcher {
	if p.OnWatch == nil {
		return nil
	}

	return p.OnWatch
}

func (p *Base) Commands() *cli.Command {
	if p.OnCommands == nil {
		return nil
	}

	return p.OnCommands()
}

func (p *Base) Flags() types.Flags {
	if p.OnFlags == nil {
		return nil
	}

	return p.OnFlags()
}
