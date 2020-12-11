package golug_codex

import (
	"sync"

	"github.com/pubgo/xerror"
)

var data sync.Map

func Register(name string, codec Codec) {
	if codec == nil {
		xerror.Next().Panic(xerror.Fmt("[codec] %s is nil", name))
	}

	data.Store(name, codec)
}

func Get(name string) Codec {
	val, ok := data.Load(name)
	if ok {
		return val.(Codec)
	}

	xerror.Next().Panic(xerror.Fmt("%s not found", name))
	return nil
}

func List() map[string]Codec {
	var dt = make(map[string]Codec)
	data.Range(func(key, value interface{}) bool {
		dt[key.(string)] = value.(Codec)
		return true
	})
	return dt
}
