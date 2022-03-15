package plugin

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/huandu/go-clone"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/spf13/cast"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/resource"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/watcher/watcher_type"
)

var _ json.Marshaler = (*Base)(nil)
var _ Plugin = (*Base)(nil)

type Base struct {
	Name         string
	Short        string
	Url          string
	Docs         interface{}
	Builder      resource.BuilderFactory
	OnHealth     types.Healthy
	OnMiddleware types.Middleware
	OnInit       func(p Process)
	OnCommands   func() *types.Command
	OnFlags      func() types.Flags
	OnWatch      func(name string, r *watcher_type.WatchResp) error
	OnVars       func(v types.Vars)

	beforeStarts []func()
	afterStarts  []func()
	beforeStops  []func()
	afterStops   []func()

	cfgMap *typex.RwMap
	cfg    config_type.IConfig
}

func (p *Base) InitCfg(i config_type.IConfig) { p.cfg = i }
func (p *Base) BeforeStart(fn func())         { p.beforeStarts = append(p.beforeStarts, fn) }
func (p *Base) AfterStart(fn func())          { p.afterStarts = append(p.afterStarts, fn) }
func (p *Base) BeforeStop(fn func())          { p.beforeStops = append(p.beforeStops, fn) }
func (p *Base) AfterStop(fn func())           { p.afterStops = append(p.afterStops, fn) }
func (p *Base) BeforeStarts() []func()        { return p.beforeStarts }
func (p *Base) AfterStarts() []func()         { return p.afterStarts }
func (p *Base) BeforeStops() []func()         { return p.beforeStops }
func (p *Base) AfterStops() []func()          { return p.afterStops }

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
	data["default_cfg"] = p.Builder.Builder()
	data["cfg"] = p.cfgMap.Map()
	data["descriptor"] = p.Short
	data["url"] = p.Url
	data["health"] = p.getFuncStack(p.OnHealth)
	data["middleware"] = p.getFuncStack(p.OnMiddleware)
	data["init"] = p.getFuncStack(p.OnInit)
	data["commands"] = p.getFuncStack(p.OnCommands)
	data["flags"] = p.getFuncStack(p.OnFlags)
	data["watch"] = p.getFuncStack(p.OnWatch)
	data["exp-var"] = p.getFuncStack(p.OnVars)

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
func (p *Base) String() string               { return p.Short }
func (p *Base) ID() string                   { return p.Name }
func (p *Base) Init() (gErr error) {
	defer xerror.Resp(func(err xerror.XErr) {
		gErr = err.WrapF("plugin: %s", p.Name)
	})

	var val = p.cfg.Get(p.Name)
	if val == nil {
		return nil
	}

	for _, data := range cast.ToSlice(val) {
		if p.cfgMap == nil {
			p.cfgMap = &typex.RwMap{}
		}

		var dm, err = cast.ToStringMapE(data)
		xerror.Panic(err)

		var base = clone.Clone(p.Builder)
		merge.MapStruct(base, dm)
		delete(dm, "_id")

		resId := base.(resource.BuilderFactory).GetResId()

		if _, ok := p.cfgMap.Load(resId); ok {
			return fmt.Errorf("res=>%s key=>%s,res key already exists", p.Name, resId)
		}

		cfg1 := clone.Clone(p.Builder.Builder())
		merge.MapStruct(cfg1, dm)
		p.cfgMap.Set(resId, cfg1)
	}

	if p.cfgMap == nil {
		p.cfgMap = &typex.RwMap{}
		cfg1 := clone.Clone(p.Builder.Builder())
		merge.MapStruct(cfg1, p.cfg.GetMap(p.Name))
		p.cfgMap.Set(consts.KeyDefault, cfg1)
	}

	// update resource
	p.cfgMap.Range(func(key string, value interface{}) bool {
		resource.Update(p.Name, key, value.(resource.BuilderFactory))
		return true
	})

	if p.OnInit != nil {
		p.OnInit(p)
		return
	}
	return nil
}

func (p *Base) Watch(name string, r *watcher_type.WatchResp) error {
	var val, ok = p.cfgMap.Load(name)
	if !ok {
		// 配置不存在
		cfg1 := clone.Clone(p.Builder.Builder())
		merge.MapStruct(cfg1, p.cfg.GetMap(p.Name, name))
		p.cfgMap.Set(name, cfg1)
	} else {
		xerror.Panic(r.Decode(val))
		p.cfgMap.Set(name, val)
	}

	resource.Update(p.Name, name, val.(resource.BuilderFactory))

	if p.OnWatch != nil {
		xerror.Panic(p.OnWatch(name, r))
	}
	return nil
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
