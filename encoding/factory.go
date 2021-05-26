package encoding

import (
	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/xerror"
)

var data typex.Map

func Register(name string, cdc Codec) {
	xerror.Assert(cdc == nil || name == "" || cdc.Name() == "", "[cdc] %s is null", name)
	xerror.Assert(data.Has(name), "[cdc] %s already exists", name)
	data.Set(name, cdc)
}

func Get(name string) Codec {
	val, ok := data.Load(name)
	if !ok {
		return nil
	}

	return val.(Codec)
}

func Keys() []string { return data.Keys() }

func Each(fn func(name string, cdc Codec)) {
	data.Each(func(name string, val interface{}) {
		fn(name, val.(Codec))
	})
}

func init() {
	vars.Watch(Name, func() interface{} { return Keys() })
}
