package golug_codec

import (
	"github.com/pubgo/golug/golug_types"
	"github.com/pubgo/xerror"
)

var data = golug_types.NewSyncMap()

func Register(name string, codec Codec) {
	xerror.Assert(codec == nil || name == "", "[codec] %s is nil", name)
	xerror.Assert(data.Has(name), "[codec] %s already exists", name)

	data.Set(name, codec)
}

func Get(name string) Codec {
	val, ok := data.Load(name)
	if !ok {
		return nil
	}

	return val.(Codec)
}

func List() (dt map[string]Codec) { data.Map(&dt); return }
