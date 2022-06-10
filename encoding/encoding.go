package encoding

import (
	"github.com/pubgo/lava/internal/pkg/typex"
	"github.com/pubgo/xerror"
)

var data typex.Map

func Register(name string, cdc Codec) {
	defer xerror.RecoverAndExit()
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
	defer xerror.RecoverAndExit()

	data.Each(func(name string, val interface{}) {
		fn(name, val.(Codec))
	})
}
