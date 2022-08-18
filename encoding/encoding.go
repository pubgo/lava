package encoding

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/internal/pkg/typex"
)

var data typex.Map

func Register(name string, cdc Codec) {
	defer recovery.Exit()
	assert.If(cdc == nil || name == "" || cdc.Name() == "", "codec[%s] is null", name)
	assert.If(data.Has(name), "[cdc] %s already exists", name)
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
	defer recovery.Exit()

	data.Each(func(name string, val interface{}) {
		fn(name, val.(Codec))
	})
}
