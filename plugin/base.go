package plugin

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/huandu/go-clone"
	"github.com/kr/pretty"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/spf13/cast"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/plugins/healthy/healthy_type"
	"github.com/pubgo/lava/resource/resource_type"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/service/service_type"
	"github.com/pubgo/lava/vars/vars_type"
	"github.com/pubgo/lava/watcher/watcher_type"
)

var _ json.Marshaler = (*Base)(nil)
var _ Plugin = (*Base)(nil)

type Base struct {
	Name           string
	Short          string
	Url            string
	Docs           interface{}
	BuilderFactory resource_type.BuilderFactory
	OnHealth       healthy_type.Handler
	OnMiddleware   service_type.Middleware
	OnInit         func(p Process)
	OnCommands     func() *typex.Command
	OnFlags        func() typex.Flags
	OnWatch        watcher_type.WatchHandler
	OnVars         func(v vars_type.Vars)

	beforeStarts []func()
	afterStarts  []func()
	beforeStops  []func()
	afterStops   []func()

	cfg    config_type.Config
	cfgMap *typex.RwMap
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
	data["default_cfg"] = p.BuilderFactory.Builder()
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

func (p *Base) Vars(f vars_type.Vars) error {
	if p.OnVars == nil {
		return nil
	}

	return xerror.Try(func() { p.OnVars(f) })
}

func (p *Base) Health() healthy_type.Handler {
	if p.OnHealth == nil {
		return nil
	}

	return p.OnHealth
}

func (p *Base) Middleware() service_type.Middleware { return p.OnMiddleware }
func (p *Base) String() string                      { return fmt.Sprintf("%s: %s", p.Name, p.Short) }
func (p *Base) ID() string                          { return p.Name }
func (p *Base) Init(cfg config_type.Config) (gErr error) {
	p.cfg = cfg

	defer xerror.Resp(func(err xerror.XErr) {
		gErr = err.WrapF("plugin: %s", p.Name)
		if runtime.IsDev() || runtime.IsTest() {
			pretty.Println(p.Name, cfg.GetMap(p.Name))
		}
	})

	var cfgVal = cfg.Get(p.Name)
	xerror.Assert(cfgVal == nil, "config(%s) not found", p.Name)

	if p.BuilderFactory != nil && p.BuilderFactory.IsValid() && cfgVal != nil {
		for _, data := range cast.ToSlice(cfgVal) {
			if p.cfgMap == nil {
				p.cfgMap = &typex.RwMap{}
			}

			var dm, err = cast.ToStringMapE(data)
			xerror.Panic(err)

			resId := resource_type.GetResId(dm)

			if _, ok := p.cfgMap.Load(resId); ok {
				return fmt.Errorf("res=>%s key=>%s,res key already exists", p.Name, resId)
			}

			resCfg := clone.Clone(p.BuilderFactory.Builder())
			merge.MapStruct(resCfg, dm)
			p.cfgMap.Set(resId, resCfg)
		}

		// if config is not slice
		if p.cfgMap == nil {
			p.cfgMap = &typex.RwMap{}
			resCfg := clone.Clone(p.BuilderFactory.Builder())
			merge.MapStruct(resCfg, cfg.GetMap(p.Name))
			p.cfgMap.Set(consts.KeyDefault, resCfg)
		}

		p.cfgMap.Range(func(key string, val interface{}) bool {
			p.BuilderFactory.Update(key, p.Name, val.(resource_type.Builder))
			return true
		})
	}

	if p.OnInit != nil {
		p.OnInit(p)
		return
	}
	return nil
}

func (p *Base) Watch(name string, r *watcher_type.Response) error {
	var val, ok = p.cfgMap.Load(name)
	var newCfg = clone.Clone(p.BuilderFactory.Builder())
	if ok {
		newCfg = clone.Clone(val)
	}
	xerror.Panic(r.Decode(newCfg))

	if !reflect.DeepEqual(val, newCfg) {
		p.BuilderFactory.Update(p.Name, name, val.(resource_type.Builder))
		p.cfgMap.Set(name, val)

		if p.OnWatch != nil {
			xerror.Panic(p.OnWatch(name, r))
		}
	}

	return nil
}

func (p *Base) Commands() *cli.Command {
	if p.OnCommands == nil {
		return nil
	}

	return p.OnCommands()
}

func (p *Base) Flags() typex.Flags {
	if p.OnFlags == nil || len(p.OnFlags()) == 0 {
		return typex.Flags{}
	}

	return p.OnFlags()
}
