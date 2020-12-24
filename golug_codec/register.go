package golug_codec

import (
	"sync"

	"github.com/pubgo/xerror"
)

var data sync.Map

func Register(name string, codec Codec) {
	if codec == nil || name == "" {
		xerror.Next().Panic(xerror.Fmt("[codec] %s is nil", name))
	}

	if _, ok := data.LoadOrStore(name, codec); ok {
		xerror.Next().Panic(xerror.Fmt("[codec] %s already exists", name))
	}
}

func Get(name string) Codec {
	val, ok := data.Load(name)
	if ok {
		return val.(Codec)
	}
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
