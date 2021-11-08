package plugin

import (
	"context"
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
	OnHealth     func(ctx context.Context) error
	OnMiddleware types.Middleware
	OnInit       func()
	OnCommands   func() *cli.Command
	OnFlags      func() []cli.Flag
	OnWatch      func(name string, resp *types.WatchResp) error
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
	data["expvar"] = p.getFuncStack(p.OnVars)
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
		return nil
	}
	return p.OnHealth
}

func (p *Base) Middleware() types.Middleware { return p.OnMiddleware }
func (p *Base) String() string               { return p.Descriptor }
func (p *Base) UniqueName() string           { return p.Name }
func (p *Base) Init() error {
	if p.OnInit == nil {
		return nil
	}

	p.OnInit()
	return nil
}

func (p *Base) Watch() func(name string, r *types.WatchResp) (err error) {
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

func (p *Base) Flags() []cli.Flag {
	if p.OnFlags == nil {
		return nil
	}

	return p.OnFlags()
}
