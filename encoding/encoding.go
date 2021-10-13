package encoding

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/typex"
)

var data typex.Map

func Register(name string, cdc Codec) {
	defer xerror.RespExit()
	xerror.Assert(cdc == nil || name == "" || cdc.Name() == "", "codec[%s] is null", name)
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
	defer xerror.RespExit()

	data.Each(func(name string, val interface{}) {
		fn(name, val.(Codec))
	})
}
