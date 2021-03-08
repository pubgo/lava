package codec

import (
	"github.com/pubgo/golug/types"
	"github.com/pubgo/xerror"
)

var data types.SMap

func List() (dt map[string]Codec) { xerror.Panic(data.Map(&dt)); return }

func Register(name string, codec Codec) {
	xerror.Assert(codec == nil || name == "" || codec.Name() == "", "[codec] %s is nil", name)
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
