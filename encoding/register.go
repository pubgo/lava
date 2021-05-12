package encoding

import (
	"github.com/pubgo/lug/types"
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/xerror"
)

var data types.SMap

func List() (dt map[string]Codec) { xerror.Panic(data.MapTo(&dt)); return }

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

func init() {
	vars.Watch(Name, func() interface{} {
		var dt []string
		xerror.Panic(data.Each(func(key string) { dt = append(dt, key) }))
		return dt
	})
}
