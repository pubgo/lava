package factory

import (
	"strings"

	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/lug/types"
	
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
)

var factories typex.SMap

type Factory struct {
	Name       string                                      `json:"name"`
	Descriptor string                                      `json:"descriptor"`
	Type       string                                      `json:"type"`
	Kind       string                                      `json:"kind"`
	Url        string                                      `json:"url"`
	Extra      types.CfgMap                                `json:"extra"`
	Handler    func(cfg types.CfgMap) (interface{}, error) `json:"-"`
	srv        interface{}
}

func (t *Factory) Get() interface{} { return t.srv }
func (t *Factory) Init(cfg types.CfgMap) error {
	if t.Handler == nil {
		return xerror.Fmt("[factory] handler is nil")
	}

	var srv, err = t.Handler(cfg)
	if err != nil {
		return xerror.Wrap(err)
	}

	t.srv = srv
	return nil
}

func Register(factory *Factory) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(factory == nil, "[factory] factory is nil")
	xerror.Assert(factory.Name == "" || factory.Type == "", "[factory] [name,type] is null")

	name := nameJoin(factory.Type, factory.Name)
	val, ok := factories.Load(name)
	xerror.Assert(ok, "[factory] already exists, name=%s, stack=%s ", name, stack.Func(val.(*Factory).Handler))

	factories.Set(name, factory)
	return
}

func Get(kind string, name string) *Factory {
	var val, ok = factories.Load(nameJoin(kind, name))
	if ok {
		return val.(*Factory)
	}

	return nil
}

func List() (fs map[string]*Factory) {
	xerror.Panic(factories.MapTo(&fs))
	return
}

func nameJoin(names ...string) string {
	return strings.Join(names, "-")
}
